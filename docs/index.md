---
layout: page
---

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vitepress'

onMounted(() => {
  const router = useRouter()
  const lang = navigator.language?.startsWith('zh') ? '/czt-contrib/zh/' : '/czt-contrib/en/'
  window.location.replace(lang)
})
</script>

<div style="text-align:center;padding:4rem">
  <p>Redirecting...</p>
  <p><a href="/czt-contrib/en/">English</a> | <a href="/czt-contrib/zh/">中文</a></p>
</div>
