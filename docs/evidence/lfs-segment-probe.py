import hashlib
import http.client
import json
import ssl
import sys
import time
import urllib.parse

auth = json.load(sys.stdin)
role = sys.argv[1]
routes = {
    'server': [('server_loopback', 'http', '127.0.0.1', 3000, '')],
    'relay': [('relay_tunnel', 'http', '127.0.0.1', 23000, ''),
              ('relay_tls_origin', 'origin', 'sub.kessokuu.top', 443, '/atlas-git')],
    'client': [('client_public_tls', 'https', 'sub.kessokuu.top', 443, '/atlas-git')],
}[role]

class OriginConnection(http.client.HTTPSConnection):
    def connect(self):
        self.sock = self._create_connection(('127.0.0.1', self.port), self.timeout)
        self.sock = self._context.wrap_socket(self.sock, server_hostname=self.host)

def connection(route):
    _, scheme, host, port, _ = route
    cls = OriginConnection if scheme == 'origin' else http.client.HTTPSConnection if scheme == 'https' else http.client.HTTPConnection
    return cls(host, port, timeout=20)

sessions = []
for route in routes:
    name, _, _, _, prefix = route
    c = connection(route)
    body = json.dumps({'operation': 'download', 'transfers': ['basic'],
                       'objects': [{'oid': auth['oid'], 'size': auth['size']}]})
    c.request('POST', prefix+'/'+auth['repo']+'.git/info/lfs/objects/batch', body,
              {'Authorization': auth['authorization'], 'Content-Type': 'application/vnd.git-lfs+json',
               'Accept': 'application/vnd.git-lfs+json', 'User-Agent': 'Atlas-segment-probe'})
    response = c.getresponse()
    data = response.read()
    if response.status != 200:
        raise SystemExit(f'{name}: batch HTTP {response.status}')
    obj = json.loads(data)['objects'][0]
    if 'error' in obj:
        raise SystemExit(f'{name}: object error {obj["error"].get("code")}')
    action = obj['actions']['download']
    path = urllib.parse.urlsplit(action['href']).path
    if not path.startswith('/'+auth['repo']+'.git/info/lfs/'):
        raise SystemExit('unexpected LFS action path')
    sessions.append((route, c, prefix+path, action.get('header', {})))

result = []
for i in range(3):
    for route, c, path, headers in (sessions if i % 2 == 0 else list(reversed(sessions))):
        started = time.perf_counter()
        c.request('GET', path, headers={**headers, 'User-Agent': 'Atlas-segment-probe'})
        response = c.getresponse()
        header_seconds = time.perf_counter()-started
        first = response.read(1)
        first_seconds = time.perf_counter()-started
        data = first+response.read()
        total = time.perf_counter()-started
        if response.status != 200 or len(data) != auth['size'] or hashlib.sha256(data).hexdigest() != auth['oid']:
            raise SystemExit(f'{route[0]}: invalid payload, HTTP {response.status}')
        result.append({'stage': route[0], 'sample': i+1, 'bytes': len(data),
                       'header_seconds': header_seconds, 'first_byte_seconds': first_seconds,
                       'total_seconds': total, 'mbps': len(data)*8/total/1e6,
                       'body_mbps': (len(data)-1)*8/(total-first_seconds)/1e6,
                       'payload_verified': True})
for _, c, _, _ in sessions:
    c.close()
print(json.dumps({'scope': 'One existing private LFS object, same byte size, persistent HTTP after authenticated batch; stages sampled separately.',
                  'role': role, 'samples': result}, indent=2))
