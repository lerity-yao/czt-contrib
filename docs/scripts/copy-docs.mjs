import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '..', '..');
const docsRoot = path.resolve(__dirname, '..');

const copyPairs = [
  { src: '../snake/README.md', dst: 'en/modules/snake.md' },
  { src: '../snake/readme-cn.md', dst: 'zh/modules/snake.md' },
  { src: '../cron/README.md', dst: 'en/modules/cron.md' },
  { src: '../cron/readme-cn.md', dst: 'zh/modules/cron.md' },
  { src: '../configcenter/consul/README.md', dst: 'en/modules/configcenter.md' },
  { src: '../configcenter/consul/readme-cn.md', dst: 'zh/modules/configcenter.md' },
  { src: '../registercenter/consul/README.md', dst: 'en/modules/registercenter.md' },
  { src: '../registercenter/consul/readme-cn.md', dst: 'zh/modules/registercenter.md' },
  { src: '../mq/rabbitmq/README.md', dst: 'en/modules/rabbitmq.md' },
  { src: '../mq/rabbitmq/readme-cn.md', dst: 'zh/modules/rabbitmq.md' },
  { src: '../aliyun/gateway/README.md', dst: 'en/modules/aliyun-gateway.md' },
  { src: '../aliyun/gateway/readme-cn.md', dst: 'zh/modules/aliyun-gateway.md' },
  { src: '../kong/hmacauth/README.md', dst: 'en/modules/kong-hmacauth.md' },
  { src: '../kong/hmacauth/readme-cn.md', dst: 'zh/modules/kong-hmacauth.md' },
  { src: '../cztctl/README.md', dst: 'en/modules/cztctl.md' },
  { src: '../cztctl/readme-cn.md', dst: 'zh/modules/cztctl.md' },
];

const changelogSections = {
  en: {
    title: '# Changelog',
    dst: 'en/changelog.md',
    sections: [
      { title: 'Snake', src: '../snake/CHANGELOG.md' },
      { title: 'Cron', src: '../cron/CHANGELOG.md' },
      { title: 'ConfigCenter (Consul)', src: '../configcenter/consul/CHANGELOG.md' },
      { title: 'RegisterCenter (Consul)', src: '../registercenter/consul/CHANGELOG.md' },
      { title: 'RabbitMQ', src: '../mq/rabbitmq/CHANGELOG.md' },
      { title: 'Aliyun Gateway', src: '../aliyun/gateway/CHANGELOG.md' },
      { title: 'Kong HMAC Auth', src: '../kong/hmacauth/CHANGELOG.md' },
      { title: 'cztctl', src: '../cztctl/CHANGELOG.md' },
    ],
  },
  zh: {
    title: '# 更新日志',
    dst: 'zh/changelog.md',
    sections: [
      { title: 'Snake', src: '../snake/changelog-cn.md' },
      { title: 'Cron', src: '../cron/changelog-cn.md' },
      { title: 'ConfigCenter (Consul)', src: '../configcenter/consul/changelog-cn.md' },
      { title: 'RegisterCenter (Consul)', src: '../registercenter/consul/changelog-cn.md' },
      { title: 'RabbitMQ', src: '../mq/rabbitmq/changelog-cn.md' },
      { title: 'Aliyun Gateway', src: '../aliyun/gateway/changelog-cn.md' },
      { title: 'Kong HMAC Auth', src: '../kong/hmacauth/changelog-cn.md' },
      { title: 'cztctl', src: '../cztctl/changelog-cn.md' },
    ],
  },
};

const gettingStarted = [
  { src: '../README.md', dst: 'en/guide/getting-started.md' },
  { src: '../readme-cn.md', dst: 'zh/guide/getting-started.md' },
];

let copied = 0;
let warnings = 0;

function ensureDir(filePath) {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
}

function copyFile(src, dst) {
  const absSrc = path.resolve(docsRoot, src);
  const absDst = path.resolve(docsRoot, dst);
  if (!fs.existsSync(absSrc)) {
    console.warn(`Warning: source not found: ${src}`);
    warnings++;
    return;
  }
  ensureDir(absDst);
  fs.copyFileSync(absSrc, absDst);
  copied++;
}

function stripFirstHeading(content) {
  return content.replace(/^\s*#\s+.*$/m, '').trim();
}

function buildChangelog(config) {
  const absDst = path.resolve(docsRoot, config.dst);
  ensureDir(absDst);
  let output = `${config.title}\n\n`;
  for (const section of config.sections) {
    const absSrc = path.resolve(docsRoot, section.src);
    if (!fs.existsSync(absSrc)) {
      console.warn(`Warning: changelog source not found: ${section.src}`);
      warnings++;
      continue;
    }
    let content = fs.readFileSync(absSrc, 'utf-8');
    content = stripFirstHeading(content);
    output += `## ${section.title}\n\n${content}\n\n`;
  }
  fs.writeFileSync(absDst, output.trimEnd() + '\n');
  copied++;
}

for (const pair of copyPairs) {
  copyFile(pair.src, pair.dst);
}

for (const key of Object.keys(changelogSections)) {
  buildChangelog(changelogSections[key]);
}

for (const pair of gettingStarted) {
  copyFile(pair.src, pair.dst);
}

console.log(`Docs copy summary: ${copied} files created/copied, ${warnings} warnings.`);
