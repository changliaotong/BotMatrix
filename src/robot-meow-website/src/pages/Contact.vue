<template>
  <section class="contact p-4">
    <h2>{{ $t('contact.title') }}</h2>
    <form @submit.prevent="onSubmit" class="form">
      <div>
        <label>{{ $t('contact.name') }}</label>
        <input type="text" v-model="name" required />
      </div>
      <div>
        <label>{{ $t('contact.email') }}</label>
        <input type="email" v-model="email" required />
      </div>
      <div>
        <label>{{ $t('contact.message') }}</label>
        <textarea v-model="message" required></textarea>
      </div>
      <div>
        <label><input type="checkbox" v-model="agree"/> {{ $t('contact.agree') }}</label>
      </div>
      <button type="submit">{{ $t('contact.submit') }}</button>
    </form>
    <p v-if="status" class="status">{{ status }}</p>
  </section>
</template>

<script setup>
import { ref } from 'vue'
const name = ref('')
const email = ref('')
const message = ref('')
const agree = ref(false)
const status = ref('')
function onSubmit(){
  if(!agree.value){ status.value = '请同意隐私条款' ; return }
  status.value = '提交成功！我们将尽快与您联系。'
  // In MVP we do not actually send data
}
</script>

<style scoped>
.contact{ padding:20px; }
.form{ display:flex; flex-direction:column; gap:12px; max-width:600px; }
input, textarea{ padding:8px; border:1px solid #ddd; border-radius:4px; width:100%; }
button{ padding:10px 14px; background:#2C8F7C; color:white; border:none; border-radius:6px; cursor:pointer; }
.status{ color:green; }
</style>
