"""Start the packaged Windows GUI executable without opening a browser.
Verify its HTTP API, cross-process instance lock and lock release after a crash.
Uses a new temporary data directory; never opens or changes user media.
"""
import argparse, json, pathlib, subprocess, tempfile, time, urllib.request

def main():
    ap = argparse.ArgumentParser(); ap.add_argument('--binary', required=True); args = ap.parse_args()
    data = pathlib.Path(tempfile.mkdtemp(prefix='prism-runtime-'))
    command = [str(pathlib.Path(args.binary).resolve()), '--no-browser', '--data-dir', str(data)]
    processes = []
    def launch(previous=''):
        proc = subprocess.Popen(command, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        processes.append(proc)
        for _ in range(200):
            path = data / 'last-session-url.txt'
            url = path.read_text() if path.exists() else ''
            if url and url != previous:
                return proc, url
            if proc.poll() is not None: raise RuntimeError('Packaged application exited before becoming ready')
            time.sleep(.05)
        raise TimeoutError('Packaged application did not become ready')
    def api(url, endpoint, body=None):
        base, token = url.split('/app/')
        request = urllib.request.Request(base+'/api/'+endpoint, data=body, headers={'Authorization': 'Bearer '+token.strip('/'), 'Content-Type': 'application/json'})
        with urllib.request.urlopen(request, timeout=10) as response: return json.load(response)
    try:
        proc, url = launch()
        assert api(url, 'config')['version'] == '2.1.0-beta.1'
        second = subprocess.run(command, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=10)
        assert second.returncode != 0, 'Second process acquired the same directory'
        proc.kill(); proc.wait(timeout=10)
        restarted, second_url = launch(url)
        assert api(second_url, 'config')['version'] == '2.1.0-beta.1'
        api(second_url, 'shutdown', b'{}'); restarted.wait(timeout=10)
        print('PASS packaged Windows GUI startup, API, duplicate rejection, crash lock release, graceful shutdown')
    finally:
        for proc in processes:
            if proc.poll() is None: proc.kill(); proc.wait(timeout=10)

if __name__ == '__main__': main()
