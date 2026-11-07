// Per-scan random mobile device and browser profile, Nigerian market only.
// The browser engine is Chromium, so the device list is the Samsung models Playwright
// ships (model tokens SM-*) plus a hand made list of the phones that lead the Nigerian
// market (Tecno, Infinix, itel, Samsung, Redmi, Nokia), with the viewport and screen the
// real device has. The user agent is rebuilt from the device model with a random browser
// and a REAL Chrome for Android release, so the version data matches what Chrome sends
// in client hints. No scanner identity anywhere in the profile.
'use strict';

const { devices } = require('playwright');

// Real Chrome for Android stable releases, from the Chrome version history API
// (majors 135-151 have dropped out of that API; refresh this table as new majors
// are released and as very old ones retire).
const CHROME_FULL = {
  118: '118.0.5993.48', 119: '119.0.6045.53', 120: '120.0.6099.43', 121: '121.0.6167.71',
  122: '122.0.6261.43', 123: '123.0.6312.40', 124: '124.0.6367.42', 125: '125.0.6422.34',
  126: '126.0.6478.40', 127: '127.0.6533.51', 128: '128.0.6613.40', 129: '129.0.6668.38',
  130: '130.0.6723.40', 131: '131.0.6778.39', 132: '132.0.6834.79', 133: '133.0.6943.39',
  134: '134.0.6998.39',
  152: '152.0.7977.84', 153: '153.0.8010.53', 154: '154.0.8037.127', 155: '155.0.8059.88',
  156: '156.0.8078.81', 157: '157.0.8103.47', 158: '158.0.8127.22',
};

const MIN_MAJOR = Math.min(...Object.keys(CHROME_FULL).map(Number));
const MAX_MAJOR = Math.max(...Object.keys(CHROME_FULL).map(Number));

// Most Android devices update Chrome within weeks, so recent majors are weighted.
const CHROME_MAJORS = [
  158, 158, 158, 158, 157, 157, 157, 157, 156, 156, 156, 156, 155, 155, 155, 155,
  154, 154, 154, 154, 152, 152, 152,
  134, 132, 130, 128, 126, 124, 122, 119, 118,
];

// (): weight => sum 100
const BROWSERS = [
  ['chrome', 55],
  ['samsung', 22],
  ['brave', 10],
  ['edge', 8],
  ['kiwi', 5],
];

// Real Samsung Internet releases and the Chromium release they shipped with.
const SAMSUNG = [
  ['28.0', 130],
  ['27.0', 125],
  ['26.0', 122],
  ['25.0', 121],
];

// Devices Playwright does not ship (token, css width, css height, dpr, Android versions).
// Model tokens are what these phones actually send in their user agent, and the list
// mirrors the Nigerian market, where Tecno, Infinix and itel lead, then Samsung, Redmi
// and Nokia. Viewport is the phone's CSS width and height (720x1600 at dpr 2 is 360x800).
const EXTRA_DEVICES = [
  // Tecno (tokens like TECNO-KF6k / TECNO KF6n)
  { model: 'TECNO KF6k', w: 360, h: 800, dpr: 2, android: ['12', '13', '14'] },
  { model: 'TECNO-KF6n', w: 360, h: 800, dpr: 2, android: ['12', '13', '14'] },
  { model: 'TECNO KF5n', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'TECNO-CK7n', w: 393, h: 873, dpr: 2.75, android: ['12', '13', '14', '15'] },
  { model: 'TECNO KG5j', w: 393, h: 873, dpr: 2.75, android: ['12', '13', '14'] },
  { model: 'TECNO-LC6', w: 360, h: 800, dpr: 2, android: ['9', '10', '11'] },
  { model: 'TECNO LA6', w: 360, h: 800, dpr: 2, android: ['9', '10', '11'] },
  { model: 'TECNO-LB7', w: 360, h: 800, dpr: 2, android: ['10', '11', '12'] },
  { model: 'TECNO LE7', w: 393, h: 873, dpr: 2.75, android: ['10', '11', '12'] },
  { model: 'TECNO KI5k', w: 360, h: 800, dpr: 2, android: ['12', '13', '14'] },
  // Infinix (tokens like Infinix X662B)
  { model: 'Infinix X662B', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'Infinix X6511B', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'Infinix X6826', w: 393, h: 873, dpr: 2.75, android: ['12', '13', '14'] },
  { model: 'Infinix X688B', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'Infinix X6516', w: 360, h: 800, dpr: 2, android: ['12', '13', '14'] },
  { model: 'Infinix X6523B', w: 360, h: 800, dpr: 2, android: ['12', '13', '14'] },
  { model: 'Infinix X6819', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'Infinix X650B', w: 360, h: 800, dpr: 2, android: ['9', '10', '11'] },
  // itel
  { model: 'itel L6006', w: 360, h: 720, dpr: 2, android: ['9', '10', '11'] },
  { model: 'itel S17', w: 360, h: 800, dpr: 2, android: ['11', '12', '13'] },
  { model: 'itel A571CL', w: 360, h: 800, dpr: 2, android: ['10', '11', '12'] },
  // Samsung models widely sold or resold in Nigeria
  { model: 'SM-A035F', w: 360, h: 800, dpr: 2.75, android: ['11', '12', '13', '14'] },
  { model: 'SM-A135F', w: 360, h: 800, dpr: 2.75, android: ['12', '13', '14'] },
  { model: 'SM-A146B', w: 393, h: 873, dpr: 2.75, android: ['12', '13', '14'] },
  { model: 'SM-A245F', w: 360, h: 800, dpr: 2.75, android: ['12', '13'] },
  { model: 'SM-A546B', w: 360, h: 800, dpr: 2.75, android: ['12', '13', '14'] },
  { model: 'SM-J600F', w: 360, h: 720, dpr: 2, android: ['9', '10'] },
  { model: 'SM-A305F', w: 360, h: 800, dpr: 2.75, android: ['10', '11', '12'] },
  // Xiaomi Redmi
  { model: '22127PN0CC', w: 360, h: 800, dpr: 2.75, android: ['12', '13', '14'] },
  { model: '2201117SG', w: 360, h: 800, dpr: 2.75, android: ['11', '12', '13'] },
  { model: 'M2101K6G', w: 393, h: 874, dpr: 2.65, android: ['11', '12', '13'] },
  // Nokia (sold in Nigeria)
  { model: 'Nokia C32', w: 360, h: 800, dpr: 2, android: ['12', '13'] },
  { model: 'Nokia G21', w: 360, h: 800, dpr: 2.5, android: ['11', '12', '13'] },
];

// Nigerian Android locale mix, all on Lagos time. Most phones carry Nigerian or common
// English; some updated phones use Yoruba or Hausa.
const LOCALES = [
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['en-NG', 'Africa/Lagos'],
  ['yo-NG', 'Africa/Lagos'],
  ['ha-NG', 'Africa/Lagos'],
];

function pick(list, r) {
  return list[Math.floor(r() * list.length)];
}

function weightedPick(pairs, r) {
  let v = r() * 100;
  for (const [id, weight] of pairs) {
    v -= weight;
    if (v <= 0) return id;
  }
  return pairs[0][0];
}

function chromeFull(major) {
  return CHROME_FULL[major] || CHROME_FULL[MAX_MAJOR];
}

// Model token from an old user agent: between '; Android x; ' and the ') ' that closes
// the first parenthesis group (some Motorola tokens contain '(' but its ')' is the one
// that ends the group).
function uaModelToken(ua) {
  const i = ua.indexOf(' Android ');
  if (i < 0) return null;
  const rest = ua.slice(i + ' Android '.length);
  const semi = rest.indexOf('; ');
  if (semi < 0) return null;
  const seg = rest.slice(semi + 2);
  const end = seg.indexOf(') ');
  if (end < 0) return null;
  return seg.slice(0, end).trim();
}

function uaAndroidVer(ua) {
  const m = /Android ([0-9.]+)/.exec(ua);
  return m ? m[1] : null;
}

// Newest Chrome an Android version can plausibly run; missing keys mean no cap.
const ANDROID_MAX_CHROME = { '4': 98, '5': 98, '6': 110, '7': 110, '8': 122, '9': 122, '10': 122 };

function chromeCapFor(androidVer) {
  return ANDROID_MAX_CHROME[String(Math.floor(Number(androidVer.split('.')[0])) || 99)] || MAX_MAJOR;
}

function chromeMajorFor(androidVer, r) {
  const cap = chromeCapFor(androidVer);
  const majors = CHROME_MAJORS.filter((m) => m <= cap);
  return pick(majors, r);
}

function chromiumPhone(d) {
  if (!d || !d.isMobile || !d.hasTouch) return null;
  if (d.defaultBrowserType && d.defaultBrowserType !== 'chromium') return null;
  const model = uaModelToken(d.userAgent);
  if (!model) return null;
  if (/Lumia|Windows Phone/i.test(d.userAgent)) return null; // strange Android claims
  const android = uaAndroidVer(d.userAgent);
  if (!android) return null;
  // Android versions that cannot run any supported Chrome build are left out.
  if (chromeCapFor(android) < MIN_MAJOR) return null;
  const viewport = d.viewport || { width: 360, height: 800 };
  const dpr = d.deviceScaleFactor || 2;
  return {
    model,
    viewport,
    deviceScaleFactor: dpr,
    screen: d.screen || { width: Math.round(viewport.width * dpr), height: Math.round(viewport.height * dpr) },
    tablet: !/Mobile/.test(d.userAgent) || viewport.width > 600,
    android,
  };
}

let POOL = null;
function devicePool() {
  if (!POOL) {
    POOL = [];
    for (const [, d] of Object.entries(devices)) {
      const p = chromiumPhone(d);
      // Only Samsung models appear in Playwright's list that the Nigerian market also has
      // (their model tokens start with SM-); Pixel, Nexus, Sony and the like are left out.
      if (p && /^SM-/.test(p.model)) POOL.push(p);
    }
    for (const e of EXTRA_DEVICES) {
      // Same rule as Playwright devices: every listed Android version must be able to run
      // one of the Chrome majors we have real release data for.
      const androids = e.android.filter((v) => chromeCapFor(v) >= MIN_MAJOR);
      if (!androids.length) continue;
      POOL.push({
        model: e.model,
        viewport: { width: e.w, height: e.h },
        deviceScaleFactor: e.dpr,
        screen: { width: Math.round(e.w * e.dpr), height: Math.round(e.h * e.dpr) },
        tablet: false,
        androidPool: androids,
      });
    }
  }
  return POOL;
}

function buildUA(model, androidVer, tablet, browserId, chromeFullVer, edgeMajor, samsungVer) {
  const head = `Mozilla/5.0 (Linux; Android ${androidVer}; ${model}) AppleWebKit/537.36 (KHTML, like Gecko)`;
  const mob = tablet ? '' : ' Mobile';
  const tail = `${mob} Safari/537.36`;
  switch (browserId) {
    case 'samsung':
      return `${head} SamsungBrowser/${samsungVer} Chrome/${chromeFullVer}${tail}`;
    case 'edge':
      return `${head} Chrome/${chromeFullVer}${tail} EdgA/${edgeMajor}.0.0.0`;
    default: // chrome, brave and kiwi send a plain Chrome user agent
      return `${head} Chrome/${chromeFullVer}${tail}`;
  }
}

/**
 * Client hints that must say the same as the user agent: the brand list, the full
 * version, the device model and the platform. This is what sec-ch-ua* request headers
 * and navigator.userAgentData come from, and the headless shell would report its own
 * HeadlessChrome numbers here.
 */
function clientHints(browserId, major, full, model, androidVer, mobile) {
  const secondBrand =
    browserId === 'samsung' ? 'Samsung Internet' :
    browserId === 'edge' ? 'Microsoft Edge' :
    browserId === 'brave' ? 'Brave' :
    'Google Chrome';
  const majorStr = String(major);
  const brands = [
    { brand: 'Chromium', version: majorStr },
    { brand: secondBrand, version: majorStr },
    { brand: 'Not A(Brand', version: '99' },
  ];
  return {
    brands,
    fullVersion: full,
    fullVersionList: [
      { brand: 'Chromium', version: full },
      { brand: secondBrand, version: full },
      { brand: 'Not A(Brand', version: '99.0.0.0' },
    ],
    platform: 'Android',
    platformVersion: `${Math.floor(Number(androidVer.split('.')[0]) || 13)}.0.0`,
    architecture: 'arm',
    bitness: '64',
    model,
    mobile,
    wow64: false,
    formFactors: ['Mobile'],
  };
}

/**
 * A random mobile profile for one scan. Every scan calls this once, so links are
 * opened from different devices, browsers, locales and time zones.
 */
function randomProfile(r = Math.random) {
  const d = pick(devicePool(), r);
  const browserId = weightedPick(BROWSERS, r);
  const androidVer = d.androidPool ? pick(d.androidPool, r) : d.android;
  let major = chromeMajorFor(androidVer, r);
  let chromeFullVer = chromeFull(major);
  let samsungVer = null;
  if (browserId === 'samsung') {
    // Samsung Internet ships its own Chromium: pick a real release and its engine.
    [samsungVer, major] = pick(SAMSUNG, r);
    chromeFullVer = chromeFull(major);
  }
  const [locale, timezoneId] = pick(LOCALES, r);
  return {
    model: d.model,
    browser: browserId,
    major,
    chromeFull: chromeFullVer,
    samsungVer,
    android: androidVer,
    userAgent: buildUA(d.model, androidVer, d.tablet, browserId, chromeFullVer, major, samsungVer),
    clientHints: clientHints(browserId, major, chromeFullVer, d.model, androidVer, !d.tablet),
    viewport: d.viewport,
    deviceScaleFactor: d.deviceScaleFactor,
    screen: d.screen,
    tablet: d.tablet,
    isMobile: true,
    hasTouch: true,
    locale,
    timezoneId,
  };
}

module.exports = { randomProfile };
