import base64
import hashlib
import http.client
import json
import pathlib
import socket
import ssl
import statistics
import subprocess
import time
import urllib.parse

host = 'sub.kessokuu.top'
origin = '45.76.168.154'
interface = 'en0'
index = socket.if_nametoindex(interface)
physical_ip = subprocess.check_output(['ipconfig', 'getifaddr', interface], text=True).strip()

class OriginHTTPS(http.client.HTTPSConnection):
    def __init__(self, physical=False):
        super().__init__(host, 443, timeout=15, context=ssl.create_default_context())
        self.physical = physical
        self.proof = {}

    def connect(self):
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(5)
        if self.physical:
            # Darwin SDK netinet/in.h: IP_BOUND_IF = 25.
            sock.setsockopt(socket.IPPROTO_IP, 25, index)
        sock.connect((origin, 443))
        self.proof = {'source_matches_physical_ipv4': sock.getsockname()[0] == physical_ip,
                      'bound_interface_matches': sock.getsockopt(socket.IPPROTO_IP, 25) == index}
        sock.settimeout(15)
        self.sock = self._context.wrap_socket(sock, server_hostname=host)

connections = {
    'cloudflare_current_tun': http.client.HTTPSConnection(host, timeout=15, context=ssl.create_default_context()),
    'origin_current_tun': OriginHTTPS(),
    'origin_physical_wifi_socket': OriginHTTPS(True),
}
result = {'scope': 'Captain Mac, same HTTPS hostname and certificate validation, existing public relay and private laptop; concurrent established HTTP/1.1 connections, rotated request order. No system routes or VPN settings changed.',
          'health': [], 'lfs': [], 'route_proof': {}, 'unavailable': {}}
for name, c in list(connections.items()):
    started = time.perf_counter()
    try:
        c.request('GET', '/atlas-api/health')
        r = c.getresponse()
        data = r.read()
        assert r.status == 200 and json.loads(data)['status'] == 'ok'
        result['health'].append({'route': name, 'kind': 'cold_https', 'seconds': time.perf_counter()-started})
        if isinstance(c, OriginHTTPS):
            result['route_proof'][name] = c.proof
    except Exception as error:
        result['unavailable'][name] = type(error).__name__+': '+str(error)
        c.close()
        del connections[name]

names = list(connections)
for i in range(10):
    order = names[i % len(names):]+names[:i % len(names)]
    for name in order:
        c = connections[name]
        started = time.perf_counter()
        c.request('GET', '/atlas-api/health')
        r = c.getresponse()
        data = r.read()
        assert r.status == 200 and json.loads(data)['status'] == 'ok'
        result['health'].append({'route': name, 'kind': 'persistent_https', 'seconds': time.perf_counter()-started})

home = pathlib.Path.home()/'.config/atlas'
settings = json.loads((home/'config.json').read_text())
git = json.loads(pathlib.Path(settings['gitea_profile']).read_text())
authorization = 'Basic '+base64.b64encode((git['username']+':'+git.get('token',git['password'])).encode()).decode()
original = pathlib.Path('.private/https-clone/transport.qa-lfs').read_bytes()
oid = hashlib.sha256(original).hexdigest()
body = json.dumps({'operation': 'download', 'transfers': ['basic'], 'objects': [{'oid': oid, 'size': len(original)}]})
actions = {}
for name in names:
    c = connections[name]
    c.request('POST', '/atlas-git/nikolastar/atlas-cli-qa.git/info/lfs/objects/batch', body,
              {'Authorization': authorization, 'Content-Type': 'application/vnd.git-lfs+json', 'Accept': 'application/vnd.git-lfs+json'})
    r = c.getresponse()
    data = r.read()
    assert r.status == 200, r.status
    action = json.loads(data)['objects'][0]['actions']['download']
    path = urllib.parse.urlsplit(action['href']).path
    assert path.startswith('/nikolastar/atlas-cli-qa.git/info/lfs/')
    actions[name] = ('/atlas-git'+path, action.get('header', {}))
for i in range(3):
    order = names[i % len(names):]+names[:i % len(names)]
    for name in order:
        c = connections[name]
        path, headers = actions[name]
        started = time.perf_counter()
        c.request('GET', path, headers=headers)
        r = c.getresponse()
        first = r.read(1)
        first_seconds = time.perf_counter()-started
        data = first+r.read()
        seconds = time.perf_counter()-started
        assert r.status == 200 and data == original
        result['lfs'].append({'route': name, 'sample': i+1, 'bytes': len(data), 'first_byte_seconds': first_seconds,
                              'seconds': seconds, 'mbps': len(data)*8/seconds/1e6, 'payload_verified': True})
for c in connections.values():
    c.close()
for name in names:
    health = [x['seconds']*1000 for x in result['health'] if x['route']==name and x['kind']=='persistent_https']
    speed = [x['mbps'] for x in result['lfs'] if x['route']==name]
    print(json.dumps({'route': name, 'health_median_ms': statistics.median(health), 'lfs_median_mbps': statistics.median(speed)}))
pathlib.Path('.private/client-route-results.json').write_text(json.dumps(result,indent=2)+'\n')
