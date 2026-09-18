#!/usr/bin/env python3
"""Browser end-to-end smoke test with synthetic fixtures (Playwright required)."""
from __future__ import annotations
import argparse, json, pathlib, shutil, subprocess, sys, time, urllib.request
from playwright.sync_api import sync_playwright, expect

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8', errors='replace')

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--binary',required=True);ap.add_argument('--fixtures',required=True);ap.add_argument('--work',required=True);ap.add_argument('--report',required=True);ap.add_argument('--screenshots',required=True);ap.add_argument('--chromium');a=ap.parse_args()
    root=pathlib.Path(a.work).resolve();root.mkdir(parents=True,exist_ok=True);fi=pathlib.Path(a.fixtures);shots=pathlib.Path(a.screenshots);shots.mkdir(parents=True,exist_ok=True)
    names={'clip.mp4':'演示视频 · 提取音频.mp4','tone.wav':'无损录音 · 24-bit.wav','image.png':'透明图片 · 测试素材.png','captions.srt':'双语字幕 · 测试素材.srt','测试_mp3.ncm':'NCM 音频 · 测试素材.ncm'}
    inputs=[]
    for old,new in names.items():
        p=root/new;shutil.copy2(fi/old,p);inputs.append(str(p))
    upload=root/'演示视频 · HEVC 输出.mp4';shutil.copy2(fi/'clip.mp4',upload)
    proc=subprocess.Popen([str(pathlib.Path(a.binary).resolve()),'--no-browser','--data-dir',str(root/'app-data'),'--ffmpeg',shutil.which('ffmpeg')],stdout=subprocess.PIPE,stderr=open(root/'app-stderr.log','w'),text=True)
    url=proc.stdout.readline().strip();assert url.startswith('http://127.0.0.1:');base=url.split('/app/')[0];token=url.split('/app/')[1].strip('/')
    tests=[];errors=[];external=[]
    def record(name,ok,detail=''):
        tests.append({'name':name,'passed':bool(ok),'detail':detail});print(('PASS ' if ok else 'FAIL ')+name,flush=True)
    def api(path):
        with urllib.request.urlopen(urllib.request.Request(base+'/api/'+path,headers={'Authorization':'Bearer '+token}),timeout=10) as r:return json.load(r)
    try:
        with sync_playwright() as p:
            browser=p.chromium.launch(headless=True,executable_path=a.chromium or None,args=['--no-sandbox'])
            context=browser.new_context(viewport={'width':1600,'height':1100},device_scale_factor=1)
            page=context.new_page();page.on('pageerror',lambda e:errors.append(str(e)));page.on('request',lambda r:external.append(r.url) if not r.url.startswith(base) and not r.url.startswith('blob:') else None)
            page.goto(url);expect(page.locator('#engineFeatures')).to_contain_text('FFprobe 已连接')
            record('Desktop application initializes',page.title()=='流光转码 · PrismTranscode')
            page.click('#addPath');page.fill('#pathText','\n'.join(inputs));page.click('#confirmPaths');expect(page.locator('#rows tr')).to_have_count(5)
            record('Batch path import with Unicode filenames',len(api('queue')['jobs'])==5)
            for name,target in [(names['tone.wav'],'flac'),(names['image.png'],'webp_lossless'),(names['captions.srt'],'vtt')]:
                page.locator('#rows tr').filter(has_text=name).locator('select').select_option(target);page.wait_for_timeout(200)
            page.set_input_files('#filePicker',str(upload));expect(page.locator('#rows tr')).to_have_count(6)
            page.locator('#rows tr').filter(has_text=upload.name).locator('select').select_option('mp4_hevc');page.wait_for_timeout(500)
            q=api('queue');targets={j['info']['name']:j['options']['target'] for j in q['jobs']}
            record('Independent per-file output format selection',targets[names['tone.wav']]=='flac' and targets[names['image.png']]=='webp_lossless' and targets[names['captions.srt']]=='vtt' and targets[upload.name]=='mp4_hevc')
            record('Browser file import writes only local cache',any('imports' in pathlib.Path(j['input']).parts for j in q['jobs']))
            page.fill('#outputDir',str(root/'outputs'));page.select_option('#workers','2');page.click('#startBtn')
            expect(page.locator('.state-completed')).to_have_count(6,timeout=90000)
            q=api('queue');record('Mixed-media batch finishes through GUI',all(j['state']=='completed' and j['result']['verified'] for j in q['jobs']))
            page.locator('#rows tr').filter(has_text=names['tone.wav']).get_by_text('详情',exact=True).click();page.wait_for_selector('#detailDialog[open]')
            record('Details show real media streams and SHA-256', 'SHA-256' in page.locator('#detailBody').inner_text() and 'pcm_s24le' in page.locator('#detailBody').inner_text())
            page.screenshot(path=str(shots/'文件详情.png'),full_page=True)
            page.click('[data-close="detailDialog"]')
            with page.expect_download() as download:page.click('#exportJson')
            result=download.value;result.save_as(root/'export.json');export=json.loads((root/'export.json').read_text(encoding='utf-8'));record('JSON export is downloadable',len(export.get('jobs',[]))==6)
            page.click('#engineBtn');page.wait_for_selector('#engineDialog[open]');record('Engine capability detection exposed', 'available_outputs' in page.locator('#engineInfo').inner_text());page.click('[data-close="engineDialog"]')
            page.wait_for_timeout(1000);page.screenshot(path=str(shots/'界面预览.png'),full_page=True)
            record('Desktop has no horizontal viewport overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'))
            page.set_viewport_size({'width':400,'height':900});page.wait_for_timeout(300);record('Mobile viewport has no page overflow',page.evaluate('document.documentElement.scrollWidth<=window.innerWidth'))
            page.screenshot(path=str(shots/'窄屏预览.png'),full_page=True)
            record('No browser JavaScript exceptions',not errors,errors)
            record('No media or UI requests to external hosts',not external,external)
            browser.close()
    finally:
        try:
            req=urllib.request.Request(base+'/api/shutdown',data=b'{}',headers={'Authorization':'Bearer '+token,'Content-Type':'application/json'});urllib.request.urlopen(req,timeout=5).close();proc.wait(timeout=10)
        except Exception:proc.kill()
        report={'schema':'prism-ui-tests/1','tested_at':time.strftime('%Y-%m-%dT%H:%M:%S%z'),'environment':sys.platform+' + Playwright Chromium; native OS dialogs not covered','tests':tests,'passed':sum(t['passed'] for t in tests),'failed':sum(not t['passed'] for t in tests)};pathlib.Path(a.report).write_text(json.dumps(report,ensure_ascii=False,indent=2),encoding='utf-8')
    return 1 if report['failed'] else 0
if __name__=='__main__':raise SystemExit(main())
