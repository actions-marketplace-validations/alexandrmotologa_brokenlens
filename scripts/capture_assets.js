const { spawn, execSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const http = require('http');
const puppeteer = require('B:/workgit/bundlelens/node_modules/puppeteer');

const PORT = 8990;
const BASE_URL = `http://127.0.0.1:${PORT}`;
const OUTPUT_DIR = path.resolve(__dirname, '../docs/images');
const FRAMES_DIR = path.resolve(__dirname, '../docs/frames_tmp');

function waitForServer(url, timeoutMs = 15000) {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    function ping() {
      http.get(url, (res) => {
        if (res.statusCode === 200) {
          resolve();
        } else {
          setTimeout(ping, 200);
        }
      }).on('error', () => {
        if (Date.now() - start > timeoutMs) {
          reject(new Error('Timeout waiting for server to start'));
        } else {
          setTimeout(ping, 200);
        }
      });
    }
    ping();
  });
}

async function capture() {
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }
  if (!fs.existsSync(FRAMES_DIR)) {
    fs.mkdirSync(FRAMES_DIR, { recursive: true });
  }

  console.log('[capture] Starting brokenlens server...');
  const serverProc = spawn(path.resolve(__dirname, '../brokenlens.exe'), [
    'serve',
    '--report', path.resolve(__dirname, '../docs/sample_report.json'),
    '--port', String(PORT)
  ], { stdio: 'inherit' });

  try {
    await waitForServer(BASE_URL);
    console.log('[capture] Server is up! Launching Puppeteer...');

    const browser = await puppeteer.launch({
      executablePath: 'C:/Program Files/Google/Chrome/Application/chrome.exe',
      headless: 'new',
      args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu']
    });

    const page = await browser.newPage();

    // 1. Desktop Dashboard Screenshot (1440x900, dpr 2)
    console.log('[capture] Capturing Desktop Dashboard (1440x900)...');
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await new Promise(r => setTimeout(r, 1200));

    const dashboardPath = path.join(OUTPUT_DIR, 'brokenlens_dashboard.png');
    await page.screenshot({ path: dashboardPath, fullPage: false });
    console.log('✓ Captured:', dashboardPath);

    // 2. Redirect Chains Tab
    console.log('[capture] Capturing Redirect Chains View...');
    const buttons = await page.$$('button');
    for (const btn of buttons) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Redirect Chains')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 500));
    const redirectsPath = path.join(OUTPUT_DIR, 'brokenlens_redirects.png');
    await page.screenshot({ path: redirectsPath, fullPage: false });
    console.log('✓ Captured:', redirectsPath);

    // 3. Mobile Responsiveness Screenshot (390x844, dpr 2 - iPhone 14/15/16)
    console.log('[capture] Capturing Mobile View (390x844)...');
    // Switch back to Detected Issues tab
    const issueBtns = await page.$$('button');
    for (const btn of issueBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Detected Issues')) {
        await btn.click();
        break;
      }
    }
    await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2, isMobile: true, hasTouch: true });
    await new Promise(r => setTimeout(r, 500));
    const mobilePath = path.join(OUTPUT_DIR, 'brokenlens_mobile.png');
    await page.screenshot({ path: mobilePath, fullPage: false });
    console.log('✓ Captured:', mobilePath);

    // 4. Capture Frames for Animated Demo GIF (1200x750, clean desktop)
    console.log('[capture] Capturing animated sequence frames...');
    await page.setViewport({ width: 1200, height: 750, deviceScaleFactor: 1.5 });
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });
    await new Promise(r => setTimeout(r, 1000));

    let frameIndex = 0;
    const captureFrame = async (count = 1) => {
      for (let i = 0; i < count; i++) {
        const frameName = `frame_${String(frameIndex++).padStart(4, '0')}.png`;
        await page.screenshot({ path: path.join(FRAMES_DIR, frameName) });
      }
    };

    // Frame sequence:
    // Hold on dashboard
    await captureFrame(6);

    // Click Redirect Chains
    const tabBtns = await page.$$('button');
    for (const btn of tabBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Redirect Chains')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 300));
    await captureFrame(8);

    // Click All Crawled URLs
    for (const btn of tabBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('All Crawled URLs')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 300));
    await captureFrame(6);

    // Click back to Detected Issues
    for (const btn of tabBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Detected Issues')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 300));
    await captureFrame(4);

    // Click "Only New Issues" filter
    const filterBtns = await page.$$('button');
    for (const btn of filterBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Only New Issues')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 300));
    await captureFrame(8);

    // Click "Only New Issues" off
    for (const btn of filterBtns) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Only New Issues')) {
        await btn.click();
        break;
      }
    }
    await new Promise(r => setTimeout(r, 300));
    await captureFrame(4);

    await browser.close();

    // 5. Build Animated Demo GIF with ffmpeg
    console.log('[capture] Compiling frames into docs/images/brokenlens_demo.gif...');
    const gifPath = path.join(OUTPUT_DIR, 'brokenlens_demo.gif');
    const ffmpegCmd = `ffmpeg -y -framerate 4 -i "${path.join(FRAMES_DIR, 'frame_%04d.png')}" -vf "fps=6,scale=1000:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=128[p];[s1][p]paletteuse=dither=bayer" "${gifPath}"`;
    execSync(ffmpegCmd, { stdio: 'inherit' });
    console.log('✓ Compiled animated demo GIF:', gifPath);

    // Clean up temporary frames
    fs.rmSync(FRAMES_DIR, { recursive: true, force: true });
    console.log('✓ Cleaned up temporary frames.');

  } finally {
    serverProc.kill('SIGINT');
    console.log('[capture] Server stopped.');
  }
}

capture().catch(err => {
  console.error('[capture] Error:', err);
  process.exit(1);
});
