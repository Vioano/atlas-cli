import ctypes
import fcntl
import http.client
import importlib.util
import json
import os
import pathlib
import statistics
import sys
import tempfile
import threading
import time

sys.path.insert(0, '/opt/personal-atlas/communication')
spec = importlib.util.spec_from_file_location('atlas_actual', '/opt/personal-atlas/communication/atlas_server.py')
atlas = importlib.util.module_from_spec(spec)
spec.loader.exec_module(atlas)

class Attr(ctypes.Structure):
    _fields_ = [('type', ctypes.c_uint32), ('size', ctypes.c_uint32),
                ('config', ctypes.c_uint64), ('sample_period', ctypes.c_uint64),
                ('sample_type', ctypes.c_uint64), ('read_format', ctypes.c_uint64),
                ('flags', ctypes.c_uint64), ('wakeup', ctypes.c_uint32),
                ('bp_type', ctypes.c_uint32), ('bp_addr', ctypes.c_uint64)]

libc = ctypes.CDLL(None, use_errno=True)
results = []
ready = threading.Condition()

class MeasuredHandler(atlas.Handler):
    def setup(self):
        super().setup()
        self.counters = []
        self.counter_error = None
        for event in (0, 1):  # PERF_COUNT_HW_CPU_CYCLES, INSTRUCTIONS
            attr = Attr(type=0, size=ctypes.sizeof(Attr), config=event, read_format=3, flags=1)
            fd = libc.syscall(298, ctypes.byref(attr), 0, -1, -1, 0)
            if fd < 0:
                self.counter_error = os.strerror(ctypes.get_errno())
                break
            self.counters.append(fd)
        for fd in self.counters:
            fcntl.ioctl(fd, 0x2403, 0)
            fcntl.ioctl(fd, 0x2400, 0)

    def finish(self):
        counter_data = []
        for fd in self.counters:
            fcntl.ioctl(fd, 0x2401, 0)
            raw = os.read(fd, 24)
            values = (ctypes.c_uint64*3).from_buffer_copy(raw)
            counter_data.append({'count': values[0], 'time_enabled_ns': values[1], 'time_running_ns': values[2]})
            os.close(fd)
        super().finish()
        with ready:
            results.append({'path': self.path, 'counters': counter_data, 'counter_error': self.counter_error})
            ready.notify_all()

with tempfile.TemporaryDirectory(prefix='atlas-cpu-probe-') as directory:
    state = pathlib.Path(directory)
    (state/'server.json').write_text(json.dumps({'public_url':'http://127.0.0.1', 'mount':None, 'min_free_bytes':0}))
    store = atlas.Store(state)
    with store.db:
        store.db.execute('INSERT INTO actors(id,name,owner) VALUES(?,?,1)', ('cpuqa','CPU fixture'))
        store.db.execute('INSERT INTO credentials(hash,actor,device,created) VALUES(?,?,?,?)', (atlas.digest('fixture-token'),'cpuqa','fixture',time.time()))
        store.db.execute('INSERT INTO projects VALUES(?,?)', ('team','Team'))
        store.db.execute('INSERT INTO grants VALUES(?,?,?)', ('cpuqa','team','member'))
    server = atlas.Server(('127.0.0.1',0), store)
    server.RequestHandlerClass = MeasuredHandler
    worker = threading.Thread(target=server.serve_forever, daemon=True)
    worker.start()
    output = {'scope':'Actual deployed Handler and Store, isolated one-actor database; one persistent worker per phase; per-thread hardware cycles including kernel execution; excludes client CPU and SSH/TLS.', 'phases':[]}
    count = 500
    for index, path in enumerate(('/health','/v1/manifest')):
        client = http.client.HTTPConnection('127.0.0.1',server.server_port,timeout=5)
        timings = []
        for _ in range(count):
            start = time.perf_counter()
            client.request('GET',path,headers={'Authorization':'Bearer fixture-token'})
            response = client.getresponse()
            response.read()
            assert response.status == 200
            timings.append((time.perf_counter()-start)*1000)
        client.close()
        with ready:
            ready.wait_for(lambda: len(results)>index,timeout=2)
        assert len(results)>index
        phase={'path':path,'requests':count,'median_ms':statistics.median(timings),'min_ms':min(timings),**results[index]}
        if len(phase['counters'])==2 and not phase['counter_error']:
            for metric, data in zip(('cycles','instructions'),phase['counters']):
                phase[metric+'_per_request']=data['count']*data['time_enabled_ns']/data['time_running_ns']/count
        output['phases'].append(phase)
    server.shutdown()
    worker.join(timeout=2)
    server.server_close()
    store.db.close()
print(json.dumps(output,indent=2))
