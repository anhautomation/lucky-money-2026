<script setup>
import { ref } from 'vue'

const id = ref('')
const account = ref('')
const phone = ref('')
const result = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  result.value = ''
  error.value = ''
  loading.value = true

  try {
    const res = await fetch('http://localhost:8080/user/submit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        id: id.value,
        account_name: account.value,
        phone: phone.value,
      }),
    })

    const data = await res.json()
    if (!res.ok) throw new Error(data.error)

    result.value = `🎉 Bạn nhận được ${data.amount}.000₫`
  } catch (e) {
    error.value = e.message || 'Có lỗi xảy ra'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-red-600 flex items-center justify-center px-4">
    <div class="bg-yellow-100 rounded-2xl shadow-xl p-8 w-full max-w-md text-center">

      <h1 class="text-3xl font-bold text-red-700 mb-2">🧧 Lì xì 2026</h1>
      <p class="text-sm text-gray-600 mb-6">
        Nhập thông tin để nhận lì xì đầu năm
      </p>

      <input v-model="id" placeholder="Mã lì xì"
        class="w-full mb-3 p-3 rounded border" />
      <input v-model="account" placeholder="Tên ngân hàng"
        class="w-full mb-3 p-3 rounded border" />
      <input v-model="phone" placeholder="Số tài khoản nhận lì xì"
        class="w-full mb-4 p-3 rounded border" />

      <button
        @click="submit"
        :disabled="loading"
        class="w-full bg-red-600 text-white py-3 rounded-xl font-bold hover:bg-red-700">
        {{ loading ? 'Đang mở bao lì xì...' : 'Mở lì xì 🧧' }}
      </button>

      <p v-if="result" class="mt-4 text-green-700 font-semibold">
        {{ result }}
      </p>
      <p v-if="error" class="mt-4 text-red-600">
        {{ error }}
      </p>

    </div>
  </div>
</template>
