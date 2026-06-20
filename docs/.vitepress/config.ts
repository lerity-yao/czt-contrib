import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/czt-contrib/',
  title: 'czt-contrib',
  description: 'Microservice Governance Components for go-zero',

  locales: {
    en: {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/getting-started' },
          {
            text: 'Modules',
            items: [
              { text: 'Snake', link: '/en/modules/snake' },
              { text: 'Cron', link: '/en/modules/cron' },
              { text: 'ConfigCenter (Consul)', link: '/en/modules/configcenter' },
              { text: 'RegisterCenter (Consul)', link: '/en/modules/registercenter' },
              { text: 'RabbitMQ', link: '/en/modules/rabbitmq' },
              { text: 'Aliyun Gateway', link: '/en/modules/aliyun-gateway' },
              { text: 'Kong HMAC Auth', link: '/en/modules/kong-hmacauth' },
              { text: 'cztctl', link: '/en/modules/cztctl' },
            ],
          },
          { text: 'Changelog', link: '/en/changelog' },
          {
            text: 'GitHub',
            link: 'https://github.com/lerity-yao/czt-contrib',
          },
        ],
        sidebar: {
          '/en/guide/': [
            {
              text: 'Guide',
              items: [{ text: 'Getting Started', link: '/en/guide/getting-started' }],
            },
          ],
          '/en/modules/': [
            {
              text: 'Modules',
              items: [
                { text: 'Snake', link: '/en/modules/snake' },
                { text: 'Cron', link: '/en/modules/cron' },
                { text: 'ConfigCenter (Consul)', link: '/en/modules/configcenter' },
                { text: 'RegisterCenter (Consul)', link: '/en/modules/registercenter' },
                { text: 'RabbitMQ', link: '/en/modules/rabbitmq' },
                { text: 'Aliyun Gateway', link: '/en/modules/aliyun-gateway' },
                { text: 'Kong HMAC Auth', link: '/en/modules/kong-hmacauth' },
                { text: 'cztctl', link: '/en/modules/cztctl' },
              ],
            },
          ],
          '/en/changelog': [
            {
              text: 'Changelog',
              items: [{ text: 'Changelog', link: '/en/changelog' }],
            },
          ],
        },
      },
    },
    zh: {
      label: '中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '指南', link: '/zh/guide/getting-started' },
          {
            text: '模块',
            items: [
              { text: 'Snake 分布式 ID', link: '/zh/modules/snake' },
              { text: 'Cron 定时任务', link: '/zh/modules/cron' },
              { text: '配置中心 (Consul)', link: '/zh/modules/configcenter' },
              { text: '注册中心 (Consul)', link: '/zh/modules/registercenter' },
              { text: 'RabbitMQ 消息队列', link: '/zh/modules/rabbitmq' },
              { text: '阿里云网关', link: '/zh/modules/aliyun-gateway' },
              { text: 'Kong HMAC 认证', link: '/zh/modules/kong-hmacauth' },
              { text: 'cztctl 代码生成', link: '/zh/modules/cztctl' },
            ],
          },
          { text: '更新日志', link: '/zh/changelog' },
          {
            text: 'GitHub',
            link: 'https://github.com/lerity-yao/czt-contrib',
          },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '指南',
              items: [{ text: '快速开始', link: '/zh/guide/getting-started' }],
            },
          ],
          '/zh/modules/': [
            {
              text: '模块',
              items: [
                { text: 'Snake 分布式 ID', link: '/zh/modules/snake' },
                { text: 'Cron 定时任务', link: '/zh/modules/cron' },
                { text: '配置中心 (Consul)', link: '/zh/modules/configcenter' },
                { text: '注册中心 (Consul)', link: '/zh/modules/registercenter' },
                { text: 'RabbitMQ 消息队列', link: '/zh/modules/rabbitmq' },
                { text: '阿里云网关', link: '/zh/modules/aliyun-gateway' },
                { text: 'Kong HMAC 认证', link: '/zh/modules/kong-hmacauth' },
                { text: 'cztctl 代码生成', link: '/zh/modules/cztctl' },
              ],
            },
          ],
          '/zh/changelog': [
            {
              text: '更新日志',
              items: [{ text: '更新日志', link: '/zh/changelog' }],
            },
          ],
        },
      },
    },
  },

  themeConfig: {
    search: {
      provider: 'local',
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/lerity-yao/czt-contrib' },
    ],
  },
})
