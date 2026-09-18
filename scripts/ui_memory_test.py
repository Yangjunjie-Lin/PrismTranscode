#!/usr/bin/env python3
"""Component/integration UI test for environments that disallow browser URLs.
Loads the real embedded HTML/CSS/JS in about:blank. A narrow bridge drives the
real localhost API using Python. Does NOT validate browser navigation/network,
native OS file dialogs, downloads or Windows execution; ui_test.py covers normal
browser navigation when run on an unrestricted local development machine.
"""
from __future__ import annotations
import argparse, base64, json, pathlib, shutil, subprocess, time, urllib.request, urllib.error
from playwright.sync_api import sync_playwright

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--binary',required=True);ap.add_argument('--fixtures',required=True);ap.add_argument('--work',required=True);ap.add_argument('--report',required=True);ap.add_argument('--screenshots',required=True);ap.add_argument('--chromium',default='/usr/bin/chromium');a=ap.parse_args()
    root=pathlib.Path(a.work).resolve();root.mkdir(parents=True,exist_ok=True);fi=pathlib.Path(a.fixtures);shots=pathlib.Path(a.screenshots);shots.mkdir(parents=True,exist_ok=True);src=pathlib.Path(__file__).resolve().parents[1]
    names={'clip.mp4':'演示视频 · 提取音频.mp4','tone.wav':'无损录音 · 24-bit.wav','image.png':'透明图片 · 测试素材.png','captions.srt':'双语字幕 · 测试素材.srt','测试_mp3.ncm':'NCM 音频 · 测试素材.ncm'}
    inputs=[]
    for old,new in names.items():
        dest=root/new;shutil.copy2(fi/old,dest);inputs.append(str(dest))
    upload=root/'演示视频 · HEVC 输出.mp4';shutil.copy2(fi/'clip.mp4',upload)
    proc=subprocess.Popen([str(pathlib.Path(a.binary).resolve()),'--no-browser','--data-dir',str(root/'app-data'),'--ffmpeg',shutil.which('ffmpeg')],stdout=subprocess.PIPE,stderr=open(root/'app-stderr.log','w'),text=True)
    url=proc.stdout.readline().strip();assert url.startswith('http://127.0.0.1:');base=url.split('/app/')[0];token=url.split('/app/')[1].strip('/')
    tests=[];errors=[]
    def record(name,ok,detail=''):
        tests.append({'name':name,'passed':bool(ok),'detail':detail});print(('PASS ' if ok else 'FAIL ')+name,flush=True)
    def api(path):
        with urllib.request.urlopen(urllib.request.Request(base+'/api/'+path,headers={'Authorization':'Bearer '+token}),timeout=30) as r:return json.load(r)
    def bridge(path,options):
        assert path.startswith('/api/') and '://' not in path and '..' not in path
        headers=options.get('headers',{});data=None
        if options.get('form'):
            boundary='PrismTestBoundaryX9251';parts=[]
            for field in options['form']:
                parts.append(('--'+boundary+'\r\n').encode())
                if 'file' in field:
                    parts.append(('Content-Disposition: form-data; name="'+field['name']+'"; filename="'+field['file']+'"\r\nContent-Type: application/octet-stream\r\n\r\n').encode());parts.append(bytes(field['data']))
                else:
                    parts.append(('Content-Disposition: form-data; name="'+field['name']+'"\r\n\r\n'+field['value']).encode())
                parts.append(b'\r\n')
            parts.append(('--'+boundary+'--\r\n').encode());data=b''.join(parts);headers['Content-Type']='multipart/form-data; boundary='+boundary
        elif options.get('body') is not None:data=options['body'].encode()
        req=urllib.request.Request(base+path,data=data,headers=headers,method=options.get('method','GET'))
        try:
            with urllib.request.urlopen(req,timeout=90) as r:return {'status':r.status,'body':r.read().decode('utf-8'),'type':r.headers.get('Content-Type','application/json')}
        except urllib.error.HTTPError as e:return {'status':e.code,'body':e.read().decode(),'type':'application/json'}
    try:
        with sync_playwright() as p:
            browser=p.chromium.launch(headless=True,executable_path=a.chromium,args=['--no-sandbox']);page=browser.new_page(viewport={'width':1600,'height':1300},device_scale_factor=1);page.on('pageerror',lambda e:errors.append(str(e)));page.expose_function('prismTestAPI',bridge)
            html=(src/'web/index.html').read_text().replace('__TOKEN__',token).replace('<link rel="stylesheet" href="/assets/style.css">','').replace('<script src="/assets/app.js"></script>','');page.set_content(html);page.add_style_tag(content=(src/'web/style.css').read_text())
            page.evaluate("""() => {const values={};Object.defineProperty(window,'localStorage',{value:{getItem:k=>values[k]??null,setItem:(k,v)=>values[k]=String(v)}});window.fetch=async(path,o={})=>{const copy={method:o.method||'GET',headers:{...o.headers}};if(o.body instanceof FormData){copy.form=[];for(const [name,v]of o.body.entries()){copy.form.push(v instanceof File?{name,file:v.name,data:Array.from(new Uint8Array(await v.arrayBuffer()))}:{name,value:v})}}else copy.body=o.body;const r=await prismTestAPI(path,copy);return new Response(r.body,{status:r.status,headers:{'Content-Type':r.type}})};}""")
            page.add_script_tag(content=(src/'web/app.js').read_text());page.wait_for_function("document.getElementById('engineFeatures').textContent.includes('FFprobe 已连接')")
            record('Real HTML/CSS/JS initialize with API bridge',page.title()=='流光转码 · PrismTranscode')
            page.click('#addPath');page.fill('#pathText','\n'.join(inputs));page.click('#confirmPaths');page.wait_for_function("document.querySelectorAll('#rows tr').length===5")
            record('Unicode batch path import from interface',len(api('queue')['jobs'])==5)
            for name,target in [(names['tone.wav'],'flac'),(names['image.png'],'webp_lossless'),(names['captions.srt'],'vtt')]:
                page.locator('#rows tr').filter(has_text=name).locator('select').select_option(target);page.wait_for_timeout(200)
            # Feed an in-memory synthetic File to the application's own file-input handler.
            page.evaluate("""async f=>{const b=Uint8Array.from(atob(f.data),c=>c.charCodeAt(0));const file=new File([b],f.name,{type:'video/mp4'});await document.getElementById('filePicker').onchange({target:{files:[file],value:''}})}""",{'name':upload.name,'data':base64.b64encode(upload.read_bytes()).decode()})
            page.wait_for_function("document.querySelectorAll('#rows tr').length===6");page.locator('#rows tr').filter(has_text=upload.name).locator('select').select_option('mp4_hevc');page.wait_for_timeout(500)
            q=api('queue');targets={j['info']['name']:j['options']['target'] for j in q['jobs']}
            record('Independent output selection is saved by backend',targets[names['tone.wav']]=='flac' and targets[names['image.png']]=='webp_lossless' and targets[names['captions.srt']]=='vtt' and targets[upload.name]=='mp4_hevc')
            record('Multipart import handler creates local cached file',any('/imports/' in j['input'] for j in q['jobs']))
            page.fill('#outputDir',str(root/'outputs'));page.select_option('#workers','2');page.click('#startBtn');page.wait_for_function("document.querySelectorAll('.state-completed').length===6",timeout=90000)
            q=api('queue');record('Six mixed-media tasks execute through UI controls',all(j['state']=='completed' and j['result']['verified'] for j in q['jobs']))
            page.locator('#rows tr').filter(has_text=names['tone.wav']).get_by_text('详情',exact=True).click();page.wait_for_selector('#detailDialog[open]');record('Details expose real source codec and result SHA-256','SHA-256' in page.locator('#detailBody').inner_text() and 'pcm_s24le' in page.locator('#detailBody').inner_text());page.screenshot(path=str(shots/'文件详情.png'),full_page=True);page.click('[data-close="detailDialog"]')
            # Check the actual export payload without triggering a blocked OS download.
            record('JSON results export contains all six jobs',len(api('export?format=json').get('jobs',[]))==6)
            page.click('#engineBtn');page.wait_for_selector('#engineDialog[open]');record('Engine dialog shows capability detection','available_outputs' in page.locator('#engineInfo').inner_text());page.click('[data-close="engineDialog"]')
            page.wait_for_timeout(5600);page.screenshot(path=str(shots/'界面预览.png'),full_page=True);record('Desktop has no page-wide horizontal overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'))
            page.set_viewport_size({'width':400,'height':900});page.wait_for_timeout(300);record('Narrow viewport has no page-wide horizontal overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'));page.screenshot(path=str(shots/'窄屏预览.png'),full_page=True)
            record('No JavaScript page exceptions',not errors,errors);browser.close()
    finally:
        try:urllib.request.urlopen(urllib.request.Request(base+'/api/shutdown',data=b'{}',headers={'Authorization':'Bearer '+token,'Content-Type':'application/json'}),timeout=5).close();proc.wait(timeout=10)
        except Exception:proc.kill()
        report={'schema':'prism-ui-component-tests/1','tested_at':time.strftime('%Y-%m-%dT%H:%M:%S%z'),'environment':'Linux Chromium, in-memory actual UI assets + Python bridge to real HTTP API. Browser URL navigation, native dialogs, downloads and Windows execution NOT validated.','tests':tests,'passed':sum(t['passed'] for t in tests),'failed':sum(not t['passed'] for t in tests)};pathlib.Path(a.report).write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
    return 1 if report['failed'] else 0
if __name__=='__main__':raise SystemExit(main())
