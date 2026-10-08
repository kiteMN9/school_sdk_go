// ==UserScript==
// @name         正方教务系统 滑块验证码自动通过（纯 HTTP 版）
// @namespace    school_sdk_go/userscript
// @version      2.0.0
// @description  直接走验证码 HTTP 接口自动通过正方教务登录页的滑块验证，不产生任何鼠标事件、不移动鼠标（算法移植自 school_sdk_go/check_code）
// @author       school_sdk_go
// @match        *://*/jwglxt*
// @match        *://*/xtgl/login_slogin.html*
// @run-at       document-idle
// @grant        none
// @noframes
// ==/UserScript==

/*
 * ============================================================================
 * 和 1.x 的区别：不再模拟鼠标拖动，整条链路全部走 HTTP 接口。
 * 拖鼠标会打断你正在用的鼠标/光标，纯 HTTP 方案完全没有这个问题。
 *
 * 流程（对应 client/login.go 的 captchaControl）：
 *   1. rtk          —— 页面上的 window.zfdun_captcha_config.rtk（拿不到就再请求一次 zfdun_captcha.js 解析 rtk:'...'）
 *   2. 取题         —— 优先复用页面上正在显示的那组图；否则 GET ?type=refresh 拿一组新的 si/mi/imtk/t
 *   3. 缺口定位     —— 用 NCC 在背景图里找拼图块的位置（见下面的算法说明）
 *   4. 造轨迹       —— 1:1 移植 check_code/gen_track.go 的“平滑插值+正弦抖动”轨迹
 *   5. 提交         —— POST ?type=verify，mt=base64(轨迹JSON)、extend=base64({appName,userAgent,appVersion})
 *   6. 判定         —— 返回 {vs:"verified", status:"success"} 即为通过（服务端已把会话标记为已验证）
 *
 * 算法说明（缺口定位）
 * ----------------------------------------------------------------------------
 * check_code/matcher.go 是“拼图块取模板 → 与自带的 10 张干净背景图滑窗取绝对差之和最小 → +70”。
 * +70 是因为那批背景图是把原始 304px 宽的背景裁掉左边 70px 得到的。
 * 这里直接抓页面上那对 304x200 的图，不需要内置背景图，也就没有 +70。
 *
 * 但不能照搬 SAD：页面上的背景图叠了一层半透明白蒙版（si = α·255 + (1-α)·原图，
 * 整图洗白、缺口处额外提亮），绝对差会被带偏 —— 实测 40 组真实样本 SAD 命中 0/40。
 * 改用对线性变换不变的归一化互相关(NCC)：正位置相关系数≈0.9999，次高峰≤0.78，40/40 命中。
 *
 * 轨迹说明
 * ----------------------------------------------------------------------------
 * check_code/gen_track.go 的时间戳是“从当前时刻往后排”的（最后一点落在未来 ~1.36s）。
 * 这里把整条轨迹平移到“提交之前已经拖完”，即最后一点 ≈ 当前时刻，
 * 与真人“拖完立刻抬手触发验证”的时间关系一致。
 * 另外轨迹坐标用的是滑块在页面上的真实位置（拿不到才退回 login.go 的 950~1030/480）。
 *
 * 实测（对真实服务端，全程无鼠标事件）：
 *   - ?type=verify 连续 10/10 返回 {vs:"verified", status:"success"}
 *   - 登录门禁验证：不做滑块验证时登录返回 #tips="请先滑动图片进行验证！"；
 *     纯 HTTP 完成滑块验证后再登录，返回 #tips="用户名或密码不正确，请重新输入！"
 *     —— 说明服务端确实放行了，滑块这一关已经过了。
 * ============================================================================
 */

(function () {
  'use strict';

  /* ============================== 配置 ============================== */
  const CFG = {
    debug: true,             // 控制台日志 + 右下角状态提示
    maxAttempts: 4,          // 一次自动通过最多尝试几轮（每轮会换一组新题）
    retryDelay: 700,         // 两次尝试之间的间隔(ms)
    cooldownMs: 1200,        // 两次自动通过之间的最小间隔，避免连续触发
    minScore: 0.95,          // NCC 峰值下限，低于此值认为没找准
    minMargin: 0.1,          // NCC 峰值与次高峰的间隔下限
    reusePageChallenge: true,// 优先复用页面上正在显示的那组图（这样界面和你看到的题一致）
    syncPageImages: true,    // 自己 refresh 换了新题后，把页面上的图也换成同一组
    markVerifiedUI: true,    // 通过后把滑块 UI 改成“验证通过”，并把拼图挪到缺口位置
    autoClickLogin: false,   // 通过后自动点“登录”（账号密码需已填好）
    useDomDragFallback: false,// 应急兜底：HTTP 始终失败时才用真实鼠标事件拖（默认关闭，避免打扰鼠标）
  };

  const TAG = '[滑块自动通过]';
  const log = (...a) => CFG.debug && console.log(TAG, ...a);
  const warn = (...a) => CFG.debug && console.warn(TAG, ...a);
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
  const randInt = (n) => Math.floor(Math.random() * n); // 等价于 Go 的 rand.Intn(n)

  let busy = false;   // 正在自动通过
  let solved = false; // 本次页面会话已通过
  let lastRun = 0;

  /* ============================== 1. HTTP 基础 ============================== */

  function getConf() {
    const c = typeof window !== 'undefined' ? window.zfdun_captcha_config : null;
    if (!c || !c.requestUrl) return null;
    return c;
  }

  const instanceOf = (conf) => conf.instanceId || 'zfcaptchaLogin';

  async function apiGet(url, accept) {
    const resp = await fetch(url, {
      credentials: 'include',
      headers: accept ? { Accept: accept } : undefined,
    });
    const text = await resp.text();
    let json = null;
    try {
      json = JSON.parse(text);
    } catch (e) {
      /* 非 JSON */
    }
    return { ok: resp.ok, status: resp.status, text: text, json: json };
  }

  /** 从 zfdun_captcha.js 资源里解析 rtk（与 client/login.go 的 getRTK 一致） */
  async function fetchRtk(conf) {
    const url = `${conf.requestUrl}?type=resource&instanceId=${encodeURIComponent(instanceOf(conf))}&name=zfdun_captcha.js`;
    const r = await apiGet(url, '*/*');
    const m = /rtk:'([^']+)'/.exec(r.text || '');
    return m ? m[1] : '';
  }

  /** GET ?type=refresh：要一组新题 */
  async function refreshChallenge(conf, rtk) {
    const url =
      `${conf.requestUrl}?type=refresh&rtk=${encodeURIComponent(rtk)}` +
      `&time=${Date.now()}&instanceId=${encodeURIComponent(instanceOf(conf))}`;
    const r = await apiGet(url, 'application/json, */*');
    return r.json || {};
  }

  /** 按页面 zfdun_captcha.js 的规则拼图片地址 */
  function imageUrl(conf, id, imtk, t) {
    return (
      `${conf.requestUrl}?type=image&id=${encodeURIComponent(id)}&imtk=${encodeURIComponent(imtk)}` +
      `&t=${t}&instanceId=${encodeURIComponent(instanceOf(conf))}`
    );
  }

  async function fetchImageBlob(url) {
    const resp = await fetch(url, { credentials: 'include' });
    if (!resp.ok) return null;
    return resp.blob();
  }

  function decodeViaImg(blob) {
    return new Promise((resolve, reject) => {
      const url = URL.createObjectURL(blob);
      const img = new Image();
      img.onload = () => {
        URL.revokeObjectURL(url);
        resolve(img);
      };
      img.onerror = () => {
        URL.revokeObjectURL(url);
        reject(new Error('图片解码失败'));
      };
      img.src = url;
    });
  }

  function decodeBlob(blob) {
    if (typeof createImageBitmap === 'function') {
      return createImageBitmap(blob).catch(() => decodeViaImg(blob));
    }
    return decodeViaImg(blob);
  }

  /** 取图像像素（自己抓的字节，画到 canvas 上不会污染，也不依赖页面上的 <img>） */
  async function fetchPixels(url) {
    const blob = await fetchImageBlob(url);
    if (!blob) return null;
    const bmp = await decodeBlob(blob);
    const w = bmp.naturalWidth || bmp.width;
    const h = bmp.naturalHeight || bmp.height;
    if (!w || !h) return null;
    const cv = document.createElement('canvas');
    cv.width = w;
    cv.height = h;
    const ctx = cv.getContext('2d', { willReadFrequently: true });
    ctx.drawImage(bmp, 0, 0, w, h);
    const img = ctx.getImageData(0, 0, w, h);
    if (typeof bmp.close === 'function') bmp.close();
    return { data: img.data, w: w, h: h };
  }

  /** POST ?type=verify（等价于 client/login.go 的 captchaVerify） */
  async function verifyTrack(conf, rtk, track) {
    const extend = b64(
      JSON.stringify({
        appName: navigator.appName,
        userAgent: navigator.userAgent,
        appVersion: navigator.appVersion,
      })
    );
    const form = new URLSearchParams({
      type: 'verify',
      rtk: rtk,
      time: String(Date.now()),
      mt: b64(JSON.stringify(track)),
      instanceId: instanceOf(conf),
      extend: extend,
    });
    const resp = await fetch(conf.requestUrl, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
        'X-Requested-With': 'XMLHttpRequest',
      },
      body: form.toString(),
    });
    const text = await resp.text();
    let json = null;
    try {
      json = JSON.parse(text);
    } catch (e) {
      /* 非 JSON */
    }
    log('verify 返回:', text.slice(0, 200));
    return json || { status: 'fail', msg: '响应不是 JSON' };
  }

  const b64 = (s) => {
    const bytes = new TextEncoder().encode(s);
    let bin = '';
    for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
    return btoa(bin);
  };

  /* ====================== 2. 轨迹生成（移植 gen_track.go） ====================== */
  /**
   * @param {number} distance 需要滑动的距离，单位是“挑战图像素”（不是页面 CSS 像素）
   * @param {number} anchorX  滑块中心的页面坐标 x
   * @param {number} anchorY  滑块中心的页面坐标 y
   * @returns {{x:number,y:number,t:number}[]} 24 个点左右、约 1.36 秒走完
   */
  function genTrack(distance, anchorX, anchorY) {
    if (distance < 70) distance = 70; // gen_track.go 的约束
    const jitter = Math.random();
    const raw = [];
    let y1 = 0;
    let current = 0;
    let t1 = 0;

    // 第一段：主要移动阶段
    while (current < distance + 3 && t1 < 1.04) {
      const base = (3 * t1 * t1 - 2 * t1 * t1 * t1) * distance * 1.008;
      const amp = Math.abs(current - distance) > 15 ? 8 : 2.5;
      current = base + Math.sin(jitter + t1 * 2) * amp;
      y1 = y1 < 5 ? y1 + randInt(2) : y1 + randInt(2) - 1;
      raw.push({ d: Math.trunc(current), dy: y1, dt: (t1 / 2.43) * 1000 });
      t1 += 0.086;
    }
    // 第二段：微调阶段
    while (t1 < 3.3) {
      const move = Math.sin(t1) * 9 + Math.sin(t1 * 2) * 1.4;
      y1 = y1 < 5 ? y1 + randInt(2) : y1 + randInt(2) - 1;
      raw.push({ d: Math.trunc(current + move), dy: y1, dt: (t1 / 2.43) * 1000 });
      t1 += 0.2;
    }

    // 位移归一化：
    //  a) 减去第一点的偏移，让轨迹从滑块起始位置 0 开始（等价于 mousedown 瞬间位移为 0）
    //  b) 缩放让最后一点正好落在缺口位置
    // 这样“末点 - 首点”正好等于缺口 x，和真人拖动产生的那条 mt 完全一致
    const d0 = raw[0].d;
    for (let i = 0; i < raw.length; i++) raw[i].d -= d0;
    const lastD = raw[raw.length - 1].d || distance;
    const k = distance / lastD;
    const tEnd = Date.now();
    const tStart = tEnd - raw[raw.length - 1].dt;
    return raw.map((p) => ({
      x: anchorX + Math.round(p.d * k),
      y: anchorY + p.dy,
      t: Math.round(tStart + p.dt),
    }));
  }

  /** 滑块在页面上的真实位置；拿不到就退回 client/login.go 用的随机起点 */
  function anchorOf() {
    const btn = document.querySelector('.captcha_wrapper .zfdun_slider_bar_btn');
    if (btn) {
      const r = btn.getBoundingClientRect();
      if (r.width > 0) {
        return {
          x: Math.round(r.left + r.width / 2 + window.scrollX),
          y: Math.round(r.top + r.height / 2 + window.scrollY),
        };
      }
    }
    return { x: 950 + randInt(81), y: 480 };
  }

  /* ============ 3. 模板提取（移植 matcher.go 的 lowerWhiteGetNonTransParentPixels） ============ */
  function buildMask(data, w, h) {
    const xs = [];
    const ys = [];
    const luma = [];
    let count = 0;
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
        count++;
        if (count % 3 === 0) continue; // 与 Go 版一致，抽样提速
        const i = (y * w + x) * 4;
        const r = data[i];
        const g = data[i + 1];
        const b = data[i + 2];
        const a = data[i + 3];
        if (r === 255 && g === 255 && b === 255 && a === 255) continue; // 纯白不透明 = 空白
        if (a > 0) {
          xs.push(x);
          ys.push(y);
          luma.push(0.299 * r + 0.587 * g + 0.114 * b);
        }
      }
    }
    return {
      n: xs.length,
      x: Int32Array.from(xs),
      y: Int32Array.from(ys),
      l: Float32Array.from(luma),
    };
  }

  /* ====================== 4. 缺口定位（NCC 滑窗匹配） ====================== */
  function matchByNCC(bg, bgW, bgH, mask, pieceW) {
    const n = mask.n;
    if (!n) return { x: 0, score: -2, margin: 0 };
    const ml = mask.l;
    const maxX = bgW - pieceW; // 与 Go 的 maxX = bounds.Dx() - dx 一致
    if (maxX < 0) return { x: 0, score: -2, margin: 0 };

    // 模板像素在背景图里的下标基数（用背景图行宽），只算一次
    const base = new Float64Array(n);
    let maxY = 0;
    for (let i = 0; i < n; i++) {
      base[i] = mask.y[i] * bgW + mask.x[i];
      if (mask.y[i] > maxY) maxY = mask.y[i];
    }
    if (maxY >= bgH) return { x: 0, score: -2, margin: 0 };

    let sx = 0;
    let sxx = 0;
    for (let i = 0; i < n; i++) {
      const v = ml[i];
      sx += v;
      sxx += v * v;
    }
    const sxxN = n * sxx - sx * sx;

    let bestX = 0;
    let best = -2;
    let second = -2;
    for (let x = 0; x <= maxX; x++) {
      let sy = 0;
      let syy = 0;
      let sxy = 0;
      for (let i = 0; i < n; i++) {
        const idx = (base[i] + x) * 4;
        const v = 0.299 * bg[idx] + 0.587 * bg[idx + 1] + 0.114 * bg[idx + 2];
        sy += v;
        syy += v * v;
        sxy += ml[i] * v;
      }
      const den = Math.sqrt(sxxN * (n * syy - sy * sy));
      const score = den === 0 ? -2 : (n * sxy - sx * sy) / den;
      if (score > best) {
        if (Math.abs(bestX - x) > 10) second = Math.max(second, best);
        best = score;
        bestX = x;
      } else if (Math.abs(x - bestX) > 10 && score > second) {
        second = score;
      }
    }
    return { x: bestX, score: best, margin: best - second };
  }

  /* ============================== 5. 页面配合 ============================== */

  /** 页面上正在显示的那组图（有的话直接复用，界面就不会和实际答的题不一致） */
  function pageImageUrls() {
    const wrapper = document.querySelector('.captcha_wrapper');
    if (!wrapper) return null;
    const bg = wrapper.querySelector('.zfdun_bgimg_img');
    const jigsaw = wrapper.querySelector('.zfdun_bgimg_jigsaw');
    const si = bg && bg.getAttribute('src');
    const mi = jigsaw && jigsaw.getAttribute('src');
    if (!si || !mi) return null;
    return { si: si, mi: mi };
  }

  /** 自己换了题之后，把页面上的图也换成同一组，保证界面一致 */
  function syncPageImages(siUrl, miUrl) {
    if (!CFG.syncPageImages) return;
    const wrapper = document.querySelector('.captcha_wrapper');
    if (!wrapper) return;
    const bg = wrapper.querySelector('.zfdun_bgimg_img');
    const jigsaw = wrapper.querySelector('.zfdun_bgimg_jigsaw');
    if (bg && bg.src !== siUrl) bg.src = siUrl;
    if (jigsaw && jigsaw.src !== miUrl) jigsaw.src = miUrl;
  }

  /** 通过后把滑块 UI 改成“验证通过”，并把拼图挪到缺口位置（纯外观，服务端状态已经好了） */
  function markVerifiedUI(gapX) {
    if (!CFG.markVerifiedUI) return;
    const wrapper = document.querySelector('.captcha_wrapper');
    if (!wrapper) return;

    if (gapX != null) {
      const bg = wrapper.querySelector('.zfdun_bgimg_img');
      const jigsaw = wrapper.querySelector('.zfdun_bgimg_jigsaw');
      let scale = 1;
      if (bg) {
        const r = bg.getBoundingClientRect();
        if (r.width > 0 && bg.naturalWidth > 0) scale = r.width / bg.naturalWidth;
      }
      if (jigsaw) jigsaw.style.left = Math.round(gapX * scale) + 'px';
    }
    const icon = wrapper.querySelector('.zfdun_slider_icon');
    if (icon) {
      icon.classList.remove('zfdun_slider_icon_normal');
      icon.classList.remove('zfdun_slider_icon_moving');
      icon.classList.add('zfdun_slider_icon_success');
    }
    const p = wrapper.querySelector('.zfdun_slider_bar p');
    if (p) p.textContent = '验证通过';
    const btn = wrapper.querySelector('.zfdun_slider_bar_btn');
    if (btn) {
      btn.classList.remove('zfdun_slider_bar_btn_fail');
      btn.style.backgroundColor = '#52CCBA';
    }
  }

  /* ============================== 6. 主流程 ============================== */
  async function solve(reason) {
    if (busy) return;
    const conf = getConf();
    if (!conf) return; // 这个页面没有滑块验证码
    if (solved && reason !== 'force') return;
    if (Date.now() - lastRun < CFG.cooldownMs) return;

    busy = true;
    lastRun = Date.now();
    try {
      const rtk = conf.rtk || (await fetchRtk(conf));
      if (!rtk) {
        warn('拿不到 rtk，放弃');
        return;
      }
      banner('识别中（纯接口，不动鼠标）…');

      let lastErr = '';
      for (let attempt = 1; attempt <= CFG.maxAttempts; attempt++) {
        let siUrl = null;
        let miUrl = null;

        // (1) 取题：优先用页面上正在显示的那组，失败/重试则自己 refresh
        if (CFG.reusePageChallenge && attempt === 1) {
          const dom = pageImageUrls();
          if (dom) {
            siUrl = dom.si;
            miUrl = dom.mi;
            log('复用页面上正在显示的这组题');
          }
        }
        if (!siUrl || !miUrl) {
          const cap = await refreshChallenge(conf, rtk);
          if (cap.vs === 'verified') {
            onSuccess(null);
            return;
          }
          if (cap.msg) log('refresh msg:', cap.msg);
          if (!cap.si || !cap.mi) {
            lastErr = cap.msg || 'refresh 没返回图片';
            await sleep(CFG.retryDelay);
            continue;
          }
          siUrl = imageUrl(conf, cap.si, cap.imtk, cap.t);
          miUrl = imageUrl(conf, cap.mi, cap.imtk, cap.t);
          syncPageImages(siUrl, miUrl);
        }

        // (2) 取像素 → 定位缺口
        const bg = await fetchPixels(siUrl);
        const piece = await fetchPixels(miUrl);
        if (!bg || !piece) {
          lastErr = '取图失败';
          await sleep(CFG.retryDelay);
          continue;
        }
        if (bg.h !== piece.h) {
          lastErr = `图片高度不一致(${bg.h}/${piece.h})`;
          await sleep(CFG.retryDelay);
          continue;
        }
        const mask = buildMask(piece.data, piece.w, piece.h);
        const res = matchByNCC(bg.data, bg.w, bg.h, mask, piece.w);
        log(
          `第${attempt}次：mask=${mask.n}px 背景=${bg.w}x${bg.h} 缺口 x=${res.x} ` +
            `相关系数=${res.score.toFixed(4)} 峰间隔=${res.margin.toFixed(4)}`
        );
        if (res.score < CFG.minScore || res.margin < CFG.minMargin) {
          lastErr = `置信度不足(score=${res.score.toFixed(3)})`;
          await sleep(CFG.retryDelay);
          continue;
        }
        if (res.x < 70) {
          // 站点实际不会出现（实测缺口 x 在 73~227 之间），出现说明识别可疑
          lastErr = `缺口位置异常 x=${res.x}`;
          await sleep(CFG.retryDelay);
          continue;
        }

        // (3) 造轨迹并提交
        const anchor = anchorOf();
        const track = genTrack(res.x, anchor.x, anchor.y);
        const ver = await verifyTrack(conf, rtk, track);
        if (ver.vs === 'verified' || ver.status === 'success') {
          onSuccess(res.x);
          return;
        }
        lastErr = ver.msg || ver.status || '未知错误';
        await sleep(CFG.retryDelay);
      }

      warn(`尝试 ${CFG.maxAttempts} 轮仍未通过：${lastErr}`);
      banner('未通过：' + lastErr);
      if (CFG.useDomDragFallback) {
        log('改用鼠标拖动兜底（CFG.useDomDragFallback 已开启）');
        await dragFallback();
      }
    } catch (e) {
      console.error(TAG, e);
    } finally {
      busy = false;
    }
  }

  function onSuccess(gapX) {
    solved = true;
    markVerifiedUI(gapX);
    banner('滑块已自动通过 ✓');
    log(gapX == null ? '服务端已是 verified，无需再验证' : `验证通过，缺口 x=${gapX}`);
    if (CFG.autoClickLogin) {
      setTimeout(() => {
        const btn = document.getElementById('dl');
        if (btn) btn.click();
      }, 300);
    }
  }

  /* ============ 7. 应急兜底：真实鼠标事件拖动（默认关闭） ============ */
  async function dragFallback() {
    const wrapper = document.querySelector('.captcha_wrapper');
    if (!wrapper) return;
    const conf = getConf();
    const cap = await refreshChallenge(conf, conf.rtk);
    if (!cap.si || !cap.mi) return;
    const siUrl = imageUrl(conf, cap.si, cap.imtk, cap.t);
    const miUrl = imageUrl(conf, cap.mi, cap.imtk, cap.t);
    syncPageImages(siUrl, miUrl);
    const bg = await fetchPixels(siUrl);
    const piece = await fetchPixels(miUrl);
    if (!bg || !piece) return;
    const res = matchByNCC(bg.data, bg.w, bg.h, buildMask(piece.data, piece.w, piece.h), piece.w);
    if (res.score < CFG.minScore) return;

    const btn = wrapper.querySelector('.zfdun_slider_bar_btn');
    if (!btn) return;
    const r = btn.getBoundingClientRect();
    const startClientX = r.left + r.width / 2;
    const startClientY = r.top + r.height / 2;
    let scale = 1;
    const bgImg = wrapper.querySelector('.zfdun_bgimg_img');
    if (bgImg) {
      const br = bgImg.getBoundingClientRect();
      if (br.width > 0 && bgImg.naturalWidth > 0) scale = br.width / bgImg.naturalWidth;
    }
    const distance = Math.round(res.x * scale);
    const track = genTrack(res.x, Math.round(startClientX + window.scrollX), Math.round(startClientY + window.scrollY));
    const rel = track.map((p) => p.x - track[0].x);
    const k = distance / (rel[rel.length - 1] || distance);

    const fire = (target, type, clientX, clientY) =>
      target.dispatchEvent(
        new MouseEvent(type, {
          bubbles: true,
          cancelable: true,
          view: window,
          clientX: Math.round(clientX),
          clientY: Math.round(clientY),
          button: 0,
          buttons: type === 'mouseup' ? 0 : 1,
        })
      );

    fire(btn, 'mousedown', startClientX, startClientY);
    const wall = performance.now();
    for (let i = 1; i < track.length; i++) {
      const wait = track[i].t - track[0].t - (performance.now() - wall);
      if (wait > 0) await sleep(wait);
      fire(window, 'mousemove', startClientX + rel[i] * k, startClientY + (track[i].y - track[0].y));
    }
    fire(window, 'mousemove', startClientX + distance, startClientY);
    fire(window, 'mouseup', startClientX + distance, startClientY);
    if (await waitFor(() => wrapper.querySelector('.zfdun_slider_icon_success'), 2500)) onSuccess(res.x);
  }

  function waitFor(fn, timeout) {
    return new Promise((resolve) => {
      const t0 = Date.now();
      (function poll() {
        let v = null;
        try {
          v = fn();
        } catch (e) {
          v = null;
        }
        if (v) return resolve(v);
        if (Date.now() - t0 > timeout) return resolve(null);
        setTimeout(poll, 50);
      })();
    });
  }

  /* ============================== 8. 状态提示 ============================== */
  function banner(text) {
    if (!CFG.debug || typeof document === 'undefined' || !document.body) return;
    let el = document.getElementById('zf-slider-captcha-tip');
    if (!el) {
      el = document.createElement('div');
      el.id = 'zf-slider-captcha-tip';
      el.style.cssText =
        'position:fixed;right:16px;bottom:16px;z-index:99999;padding:6px 10px;border-radius:4px;' +
        'background:rgba(0,0,0,.72);color:#fff;font-size:12px;line-height:1.4;pointer-events:none;' +
        'transition:opacity .4s;opacity:0;font-family:inherit;';
      document.body.appendChild(el);
    }
    el.textContent = '滑块助手：' + text;
    el.style.opacity = '1';
    clearTimeout(el._timer);
    el._timer = setTimeout(() => {
      el.style.opacity = '0';
    }, 3000);
  }

  /* ============================== 9. 启动 ============================== */
  function boot() {
    log('已加载 v2.0.0（纯 HTTP 模式，全程不会产生鼠标事件）');

    // 等滑块控件渲染出来
    let tries = 0;
    const iv = setInterval(() => {
      if (document.querySelector('.captcha_wrapper')) {
        clearInterval(iv);
        solve('boot');
        return;
      }
      if (++tries > 40) clearInterval(iv); // 10 秒还没有就不等了
    }, 250);

    // 用户点“刷新”换题时，重新走一遍
    document.addEventListener(
      'click',
      (e) => {
        const t = e.target;
        if (t && t.closest && t.closest('.zfdun_refresh_btn')) {
          solved = false;
          setTimeout(() => solve('refresh'), 500);
        }
      },
      true
    );

    // 题过期 / 服务端又要求验证时（例如点了登录被拦），自动补一次
    const obs = new MutationObserver(() => {
      if (busy) return;
      const tips = document.getElementById('tips');
      if (tips && /请先滑动图片|请拖动滑块|滑块/.test(tips.textContent || '')) {
        solved = false;
        solve('tips');
      }
    });
    obs.observe(document.documentElement, { subtree: true, childList: true, characterData: true });
  }

  /* ====================== 浏览器/测试 双端导出 ====================== */
  const api = {
    CFG,
    genTrack,
    buildMask,
    matchByNCC,
    fetchRtk,
    refreshChallenge,
    imageUrl,
    fetchPixels,
    verifyTrack,
    solve,
  };
  if (typeof window !== 'undefined') window.__zfSliderCaptcha = api;
  if (typeof document !== 'undefined' && document.querySelector) boot();
})();
