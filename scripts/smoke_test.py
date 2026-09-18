#!/usr/bin/env python3
"""Real-engine integration tests. Generates synthetic media and uses the local HTTP API.
Requires Python standard library, Go (fixture generator), FFmpeg + ffprobe.
"""
from __future__ import annotations
import argparse, hashlib, json, pathlib, shutil, subprocess, sys, time, urllib.request, urllib.error

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8', errors='replace')

def run(args, **kw):
    p=subprocess.run([str(x) for x in args],stdout=subprocess.PIPE,stderr=subprocess.PIPE,**kw)
    if p.returncode:raise RuntimeError(p.stderr.decode('utf-8','replace')[-6000:])
    return p.stdout

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--binary',required=True);ap.add_argument('--work',required=True);ap.add_argument('--report');ap.add_argument('--sample-ncm');a=ap.parse_args()
    root=pathlib.Path(a.work).resolve();root.mkdir(parents=True,exist_ok=True);fixtures=root/'fixtures';fixtures.mkdir(exist_ok=True)
    source=pathlib.Path(__file__).resolve().parents[1];binary=str(pathlib.Path(a.binary).resolve());ff=shutil.which('ffmpeg');probe=shutil.which('ffprobe');assert ff and probe
    report={'schema':'prism-smoke/1','tested_at':time.strftime('%Y-%m-%dT%H:%M:%S%z'),'platform':sys.platform,'ffmpeg':run([ff,'-version']).decode().splitlines()[0],'results':[]}
    def record(name,passed,detail=''):
        report['results'].append({'name':name,'passed':bool(passed),'detail':detail});print(('PASS ' if passed else 'FAIL ')+name+(' '+str(detail)[:160] if detail else ''),flush=True)
    fb=[ff,'-hide_banner','-loglevel','error','-nostdin','-y','-threads','2','-filter_threads','1']
    run(fb+['-f','lavfi','-i','aevalsrc=0.25*sin(2*PI*440*t)+0.001*sin(2*PI*1733*t):s=48000:d=1.2','-c:a','pcm_s24le',fixtures/'tone.wav'])
    raw24=run([ff,'-v','error','-i',fixtures/'tone.wav','-c:a','pcm_s24le','-f','s24le','-']);record('Fixture contains real low-order 24-bit sample information',any(raw24[::3]))
    run(fb+['-f','lavfi','-i','testsrc2=size=320x180:rate=25:duration=1.2','-f','lavfi','-i','sine=frequency=620:sample_rate=48000:duration=1.2','-c:v','libx264','-preset','ultrafast','-pix_fmt','yuv420p','-c:a','aac','-shortest','-threads','2',fixtures/'clip.mp4'])
    run(fb+['-f','lavfi','-i','color=c=red@0.4:s=160x90,format=rgba','-frames:v','1','-update','1',fixtures/'image.png'])
    (fixtures/'captions.srt').write_text('1\n00:00:00,000 --> 00:00:00,500\n流光转码测试\n\n2\n00:00:00,600 --> 00:00:01,100\nSynthetic subtitle.\n',encoding='utf-8')
    run(['go','run','./cmd/fixtures','-out',fixtures],cwd=source)
    run(fb+['-i',fixtures/'clip.mp4','-i',fixtures/'captions.srt','-map','0','-map','1','-c','copy',fixtures/'with_subtitle.mkv'])
    shutil.copy2(fixtures/'clip.mp4',fixtures/'misnamed.data');(fixtures/'broken.mp4').write_bytes(b'not a valid movie')
    proc=None;base=token=''
    def launch():
        nonlocal proc,base,token
        proc=subprocess.Popen([binary,'--no-browser','--data-dir',str(root/'app-data'),'--ffmpeg',ff],stdout=subprocess.PIPE,stderr=open(root/'app-stderr.txt','ab'),text=True)
        url=proc.stdout.readline().strip();assert url.startswith('http://127.0.0.1:'),url;base=url.split('/app/')[0];token=url.split('/app/')[1].strip('/')
    def api(path,body=None):
        headers={'Authorization':'Bearer '+token};data=None
        if body is not None:data=json.dumps(body).encode();headers['Content-Type']='application/json'
        req=urllib.request.Request(base+'/api/'+path,data=data,headers=headers)
        try:
            with urllib.request.urlopen(req,timeout=90) as r:return json.load(r)
        except urllib.error.HTTPError as e:raise RuntimeError(e.read().decode()) from e
    def wait(timeout=180):
        start=time.monotonic()
        while True:
            q=api('queue')
            if not q['busy']:return q
            if time.monotonic()-start>timeout:api('stop',{});raise TimeoutError('conversion timeout')
            time.sleep(.10)
    def find(path):return next(j for j in api('queue')['jobs'] if j['input']==str(path))
    def convert(path,target,extra=None,out=None):
        j=find(path);o={**defaults,'target':target,**(extra or {})};api('configure',{'ids':[j['id']],'options':o});api('start',{'ids':[j['id']],'output_dir':str(out or root/'outputs'/target),'workers':1});return next(v for v in wait()['jobs'] if v['id']==j['id'])
    def hashes(path,selector):
        raw=run([probe,'-v','error','-select_streams',selector,'-show_packets','-show_data_hash','sha256','-show_entries','packet=data_hash','-of','json',path]);return [p.get('data_hash') for p in json.loads(raw)['packets']]
    try:
        launch();c=api('config');defaults=c['defaults'];record('38 output presets',len(c['targets'])==38)
        paths=[fixtures/x for x in ['tone.wav','clip.mp4','image.png','captions.srt','测试_mp3.ncm','测试_flac.ncm','with_subtitle.mkv','misnamed.data','broken.mp4']]
        api('add',{'paths':[str(p) for p in paths],'options':defaults});record('Content detection despite misleading extension',find(fixtures/'misnamed.data')['info']['kind']=='video');outputs={}
        for t in c['targets']:
            if not t['available']:record('output '+t['id'],False,'encoder/muxer not installed');continue
            path={'audio':fixtures/'tone.wav','video':fixtures/'clip.mp4','image':fixtures/'image.png','subtitle':fixtures/'with_subtitle.mkv'}[t['kind']]
            if t['id'] in ('mp3','gif'):path=fixtures/'clip.mp4'
            if t['id']=='copy_audio':path=fixtures/'测试_mp3.ncm'
            try:
                j=convert(path,t['id']);ok=j['state']=='completed';detail=j.get('error','')
                if ok:
                    r=j['result'];outputs[t['id']]=r['output'];actual=hashlib.sha256(pathlib.Path(r['output']).read_bytes()).hexdigest();ok=r['verified'] and r['sha256']==actual;detail={'mode':r['mode'],'bytes':r['size'],'verification':r['verification'],'sha256':r['sha256']}
                record('output '+t['id'],ok,detail)
            except Exception as ex:record('output '+t['id'],False,str(ex))
        pcm=lambda p:run([ff,'-v','error','-i',p,'-map','0:a:0','-c:a','pcm_s32le','-f','s32le','-'])
        pixels=lambda p:run([ff,'-v','error','-threads','1','-i',p,'-f','rawvideo','-pix_fmt','rgba','-threads','1','-'])
        if 'flac' in outputs:record('24-bit WAV to FLAC: decoded PCM exact equality',pcm(fixtures/'tone.wav')==pcm(outputs['flac']))
        if 'tiff' in outputs:record('RGBA PNG to TIFF: decoded pixels exact equality',pixels(fixtures/'image.png')==pixels(outputs['tiff']))
        if 'png' in outputs:record('RGBA PNG to PNG: decoded pixels exact equality',pixels(fixtures/'image.png')==pixels(outputs['png']))
        if 'webp_lossless' in outputs:record('RGBA PNG to lossless WebP: decoded pixels exact equality',pixels(fixtures/'image.png')==pixels(outputs['webp_lossless']))
        if 'avif' in outputs:
            api('add',{'paths':[outputs['avif']], 'options':{**defaults,'target':'png'}});record('AVIF content is correctly classified as image',find(pathlib.Path(outputs['avif']))['info']['kind']=='image')
        if 'remux_mkv' in outputs:
            record('MP4 to MKV remux: video packets exact equality',hashes(fixtures/'clip.mp4','v:0')==hashes(outputs['remux_mkv'],'v:0'))
            record('MP4 to MKV remux: audio packets exact equality',hashes(fixtures/'clip.mp4','a:0')==hashes(outputs['remux_mkv'],'a:0'))
        if 'copy_audio' in outputs:record('NCM MP3 extraction: audio packets exact equality',hashes(source/'testdata/tone.mp3','a:0')==hashes(outputs['copy_audio'],'a:0'))
        for f,t in [('测试_mp3.ncm','mp3'),('测试_flac.ncm','flac'),('测试_flac.ncm','mp3')]:
            j=convert(fixtures/f,t);record(f+' -> '+t,j['state']=='completed',j.get('error',''))
        j=convert(fixtures/'tone.wav','wav',{'sample_rate':44100},root/'resample');record('SoXR high-precision resampling',j['state']=='completed' and 'soxr:precision=28' in ' '.join(j.get('result',{}).get('command',[])),j.get('error',''))
        orig=hashlib.sha256((fixtures/'tone.wav').read_bytes()).hexdigest();j=convert(fixtures/'tone.wav','wav',out=fixtures);record('Never overwrite original, including same extension',j['state']=='completed' and j['result']['output']!=str(fixtures/'tone.wav') and hashlib.sha256((fixtures/'tone.wav').read_bytes()).hexdigest()==orig)
        j=convert(fixtures/'tone.wav','wav',{'collision':'skip'},fixtures);record('Skip collision policy',j['state']=='skipped')
        j=convert(fixtures/'clip.mp4','mp4',{'hardware':'nvenc','fallback':True},root/'hw-fallback');record('NVENC unavailable -> disclosed CPU fallback',j['state']=='completed' and any('硬件' in w for w in j.get('result',{}).get('warnings',[])),j.get('error',''))
        bad=find(fixtures/'broken.mp4');good=find(fixtures/'tone.wav');api('configure',{'ids':[bad['id'],good['id']],'options':defaults});api('start',{'ids':[bad['id'],good['id']],'output_dir':str(root/'batch-isolation'),'workers':2});states={j['id']:j['state'] for j in wait()['jobs']};record('Damaged input does not abort parallel batch',states[bad['id']]=='failed' and states[good['id']]=='completed')
        longfile=fixtures/'long.mp4';run(fb+['-f','lavfi','-i','testsrc2=size=1280x720:rate=25:duration=8','-c:v','libx264','-preset','ultrafast','-threads','2',longfile]);api('add',{'paths':[str(longfile)],'options':{**defaults,'target':'mp4_av1'}});j=find(longfile);api('start',{'ids':[j['id']],'output_dir':str(root/'cancel-output'),'workers':1});time.sleep(.3);api('stop',{});j=next(x for x in wait()['jobs'] if x['id']==j['id']);record('Cancellation removes uncommitted output',j['state']=='cancelled' and not list((root/'cancel-output').glob('*')),j.get('error',''))
        if a.sample_ncm:
            path=pathlib.Path(a.sample_ncm).resolve();before=hashlib.sha256(path.read_bytes()).hexdigest();api('add',{'paths':[str(path)],'options':defaults});j=convert(path,'mp3',out=root/'real-sample');record('User supplied NCM real-file regression',j['state']=='completed' and before==hashlib.sha256(path.read_bytes()).hexdigest(),{'mode':j.get('result',{}).get('mode'),'duration':j.get('result',{}).get('output_info',{}).get('duration'),'original_unchanged':before==hashlib.sha256(path.read_bytes()).hexdigest(),'error':j.get('error','')})
        before_ids={j['id'] for j in api('queue')['jobs']};api('shutdown',{});proc.wait(timeout=10);launch();record('Queue restored after process restart',before_ids=={j['id'] for j in api('queue')['jobs']})
    finally:
        if proc and proc.poll() is None:
            try:api('shutdown',{});proc.wait(timeout=10)
            except Exception:proc.kill()
        report['passed']=sum(x['passed'] for x in report['results']);report['total']=len(report['results']);report['failed']=report['total']-report['passed'];rp=pathlib.Path(a.report) if a.report else root/'smoke-report.json';rp.parent.mkdir(parents=True,exist_ok=True);rp.write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8');print('REPORT',rp,flush=True)
    return 1 if report['failed'] else 0
if __name__=='__main__':sys.exit(main())
