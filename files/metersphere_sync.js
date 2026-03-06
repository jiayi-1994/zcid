const MS_BASE_URL = process.env.MS_BASE_URL ?? 'http://172.20.18.75:8081';
const MS_USERNAME = process.env.MS_USERNAME ?? '';
const MS_PASSWORD = process.env.MS_PASSWORD ?? '';
const MS_AUTH = process.env.MS_AUTH ?? 'LOCAL';
const MS_PROJECT_ID = process.env.MS_PROJECT_ID ?? '';
const MS_PROJECT_NAME = process.env.MS_PROJECT_NAME ?? '';

if (!MS_USERNAME || !MS_PASSWORD) {
  console.error('Missing credentials: set MS_USERNAME and MS_PASSWORD');
  process.exit(1);
}

class CookieJar {
  #cookies = new Map();

  absorb(headers) {
    const raw = headers.get('set-cookie');
    if (!raw) return;
    const parts = raw.split(/,(?=[^;]+?=)/g);
    for (const p of parts) {
      const kv = p.split(';', 1)[0];
      const idx = kv.indexOf('=');
      if (idx <= 0) continue;
      const name = kv.slice(0, idx).trim();
      const value = kv.slice(idx + 1).trim();
      if (!name) continue;
      this.#cookies.set(name, value);
    }
  }

  headerValue() {
    return [...this.#cookies.entries()].map(([k, v]) => `${k}=${v}`).join('; ');
  }
}

function toPemPublicKey(base64Der) {
  const cleaned = String(base64Der ?? '').trim();
  const wrapped = cleaned.replace(/(.{64})/g, '$1\n');
  return `-----BEGIN PUBLIC KEY-----\n${wrapped}\n-----END PUBLIC KEY-----`;
}

function encryptWithPublicKey(publicKeyBase64Der, plaintext) {
  const crypto = require('node:crypto');
  const pem = toPemPublicKey(publicKeyBase64Der);
  const buf = Buffer.from(String(plaintext), 'utf8');
  const enc = crypto.publicEncrypt(
    {
      key: pem,
      padding: crypto.constants.RSA_PKCS1_PADDING,
    },
    buf,
  );
  return enc.toString('base64');
}

async function msFetch(jar, path, opts = {}) {
  const url = `${MS_BASE_URL}${path}`;
  const headers = new Headers(opts.headers ?? {});
  headers.set('Accept', 'application/json');

  const cookie = jar.headerValue();
  if (cookie) headers.set('Cookie', cookie);

  if (opts.csrfToken) headers.set('CSRF-TOKEN', opts.csrfToken);
  if (opts.workspaceId) headers.set('WORKSPACE', opts.workspaceId);
  if (opts.projectId) headers.set('PROJECT', opts.projectId);

  const init = {
    method: opts.method ?? 'GET',
    headers,
    redirect: 'manual',
  };

  if (Object.prototype.hasOwnProperty.call(opts, 'json')) {
    headers.set('Content-Type', 'application/json');
    init.body = JSON.stringify(opts.json);
  }

  const res = await fetch(url, init);
  jar.absorb(res.headers);

  const contentType = res.headers.get('content-type') ?? '';
  const text = await res.text();
  const isJson = contentType.includes('application/json') || text.startsWith('{') || text.startsWith('[');

  return {
    status: res.status,
    headers: res.headers,
    bodyText: text,
    bodyJson: isJson ? safeJsonParse(text) : undefined,
    redirected: res.status >= 300 && res.status < 400,
    location: res.headers.get('location') ?? undefined,
  };
}

function findFirstString(obj, keys) {
  if (!obj || typeof obj !== 'object') return undefined;
  for (const k of keys) {
    const v = obj[k];
    if (typeof v === 'string' && v.trim()) return v;
  }
  return undefined;
}

function deepFindFirstString(obj, keys, maxDepth = 4) {
  const seen = new Set();
  const queue = [{ value: obj, depth: 0 }];
  while (queue.length) {
    const { value, depth } = queue.shift();
    if (!value || typeof value !== 'object') continue;
    if (seen.has(value)) continue;
    seen.add(value);

    const direct = findFirstString(value, keys);
    if (direct) return direct;
    if (depth >= maxDepth) continue;

    if (Array.isArray(value)) {
      for (const it of value) queue.push({ value: it, depth: depth + 1 });
      continue;
    }
    for (const k of Object.keys(value)) {
      queue.push({ value: value[k], depth: depth + 1 });
    }
  }
  return undefined;
}

function safeJsonParse(s) {
  try {
    return JSON.parse(s);
  } catch {
    return undefined;
  }
}

async function login(jar) {
  const pub = await msFetch(jar, '/isLogin', { method: 'GET' });
  const pubJson = pub.bodyJson;
  const publicKey = typeof pubJson?.message === 'string' ? pubJson.message : undefined;
  const username = publicKey ? encryptWithPublicKey(publicKey, MS_USERNAME) : MS_USERNAME;
  const password = publicKey ? encryptWithPublicKey(publicKey, MS_PASSWORD) : MS_PASSWORD;

  const res = await msFetch(jar, '/signin', {
    method: 'POST',
    json: { username, password, authenticate: MS_AUTH },
  });

  if (res.status !== 200) {
    const maybe = safeJsonParse(res.bodyText);
    const msg = typeof maybe?.message === 'string' ? maybe.message : res.bodyText;
    throw new Error(`signin failed: status=${res.status} message=${msg}`);
  }

  const body = res.bodyJson;
  if (!body || typeof body !== 'object') {
    throw new Error(`signin returned non-json: ${res.bodyText}`);
  }
  if (!body.success) {
    throw new Error(`signin unsuccessful: ${body.message}`);
  }

  const csrfToken = deepFindFirstString(body, ['csrfToken', 'csrf', 'xsrfToken', 'XSRF-TOKEN']);
  const user = body.data?.user ?? body.data ?? undefined;

  return { csrfToken, user };
}

async function main() {
  const jar = new CookieJar();
  const { csrfToken } = await login(jar);
  if (!csrfToken) {
    console.warn('Warning: csrfToken not found in /signin response; you may need to capture it from UI/localStorage.');
  }

  const cur = await msFetch(jar, '/currentUser', {
    csrfToken: csrfToken ?? '1',
  });
  if (cur.status !== 200) {
    throw new Error(`currentUser failed: status=${cur.status} body=${cur.bodyText}`);
  }

  const projects = await msFetch(jar, '/project/listAll', {
    csrfToken: csrfToken ?? '1',
  });
  if (projects.status !== 200) {
    throw new Error(`project/listAll failed: status=${projects.status} body=${projects.bodyText}`);
  }
  const projectList = Array.isArray(projects.bodyJson) ? projects.bodyJson : [];
  if (projectList.length === 0) {
    console.warn('No projects found in response; inspect body:', projects.bodyText.slice(0, 500));
  }

  const chosen =
    (MS_PROJECT_ID ? projectList.find((p) => p?.id === MS_PROJECT_ID) : undefined) ??
    (MS_PROJECT_NAME ? projectList.find((p) => p?.name === MS_PROJECT_NAME) : undefined) ??
    projectList[0];

  const projectId = chosen?.id;
  const workspaceId = chosen?.workspaceId;
  if (!projectId) {
    throw new Error('Cannot determine projectId from /project/listAll; choose a project explicitly.');
  }

  const nodes = await msFetch(jar, `/case/node/list/${encodeURIComponent(projectId)}`, {
    csrfToken: csrfToken ?? '1',
    workspaceId,
    projectId,
  });
  if (nodes.status !== 200) {
    throw new Error(`case/node/list failed: status=${nodes.status} body=${nodes.bodyText}`);
  }
  const nodeList = Array.isArray(nodes.bodyJson) ? nodes.bodyJson : [];
  const nodeId = nodeList[0]?.id;
  if (!nodeId) {
    throw new Error('Cannot determine nodeId from /case/node/list/{projectId}');
  }

  const tc = await msFetch(jar, '/test/case/save', {
    method: 'POST',
    csrfToken: csrfToken ?? '1',
    workspaceId,
    projectId,
    json: {
      projectId,
      nodeId,
      name: 'API创建用例-示例',
      priority: 'P0',
      status: 'Underway',
      tags: 'api,sync',
      prerequisite: '已登录',
      steps: '1. 调用 /signin\n2. 调用 /test/case/save',
      expectedResult: '用例创建成功',
    },
  });

  if (tc.status !== 200) {
    throw new Error(`test/case/save failed: status=${tc.status} body=${tc.bodyText}`);
  }
  console.log('Created test case:', tc.bodyText.slice(0, 500));
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});

process.on('unhandledRejection', (e) => {
  console.error(e);
  process.exitCode = 1;
});

process.on('uncaughtException', (e) => {
  console.error(e);
  process.exitCode = 1;
});
