<script setup>
import { ref } from 'vue'

const id = ref('')
const account = ref('')
const bank = ref('')
const bankno = ref('')
const fullname = ref('')
const result = ref('')
const error = ref('')
const loading = ref(false)

const isOpening = ref(false)
const showMoneyFly = ref(false)

async function submit() {
  result.value = ''
  error.value = ''

  if (!id.value || !account.value || !bank.value || !bankno.value || !fullname.value) {
    error.value = 'Vui lòng nhập đầy đủ thông tin trước khi mở lì xì'
    return
  }

  loading.value = true
  isOpening.value = true

  try {
    const res = await fetch(
      'https://lucky-money-2026-production.up.railway.app/user/submit',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: id.value,
          account: account.value,
          bank: bank.value,
          bankno: bankno.value,
          name: fullname.value
        }),
      }
    )

    const data = await res.json()
    if (!res.ok) throw new Error(data.error)

    result.value = `🎉 Bạn nhận được ${data.amount}.000₫`

    showMoneyFly.value = true
    setTimeout(() => {
      showMoneyFly.value = false
    }, 1800)

  } catch (e) {
    error.value = e.message || 'Có lỗi xảy ra'
  } finally {
    loading.value = false
    setTimeout(() => {
      isOpening.value = false
    }, 1500)
  }
}
</script>

<template>
  <div class="user-page">
    <div class="card text-center">

      <h1 class="text-3xl font-bold text-red-700 mb-2">🧧 Lì xì 2026</h1>
      <p class="text-sm text-gray-600 mb-6">
        Nhập thông tin để nhận lì xì đầu năm
      </p>

      <input
        v-model="id"
        placeholder="Mã lì xì"
        class="w-full mb-3 p-3 rounded border"
      />
      <input
        v-model="account"
        placeholder="Tên ingame"
        class="w-full mb-3 p-3 rounded border"
      />
      <input
        v-model="bank"
        placeholder="Tên ngân hàng nhận lì xì"
        class="w-full mb-3 p-3 rounded border"
      />
      <input
        v-model="bankno"
        placeholder="Số tài khoản nhận lì xì"
        class="w-full mb-4 p-3 rounded border"
      />

      <input
        v-model="fullname"
        placeholder="Thông tin họ tên"
        class="w-full mb-4 p-3 rounded border"
      />

      <button
        @click="submit"
        :disabled="loading"
        class="w-full bg-red-600 text-white py-3 rounded-xl font-bold hover:bg-red-700"
      >
        {{ loading ? 'Đang mở bao lì xì...' : 'Mở lì xì 🧧' }}
      </button>

      <p v-if="result" class="mt-4 text-green-700 font-semibold">
        {{ result }}
      </p>
      <p v-if="error" class="mt-4 text-red-600">
        {{ error }}
      </p>

      <div class="lixi-row">
        <div class="lixi" :class="{ open: isOpening }">🧧</div>
        <div class="lixi" :class="{ open: isOpening }">🧧</div>
        <div class="lixi" :class="{ open: isOpening }">🧧</div>
        <div class="lixi" :class="{ open: isOpening }">🧧</div>
      </div>

      <div v-if="showMoneyFly" class="money-fly">
        <span>💸</span>
        <span>💸</span>
        <span>💸</span>
        <span>💸</span>
        <span>💸</span>
      </div>

    </div>
  </div>
</template>

<style scoped>
.user-page {
  min-height: 100vh;
  background-image: url('/4261727762388106724.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  display: flex;
  align-items: center;
  justify-content: center;
}

.card {
  background: rgba(255, 248, 220, 0.96);
  border-radius: 20px;
  padding: 28px;
  width: 360px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.25);
}

.lixi-row {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 16px;
}

.lixi {
  font-size: 72px;
  transition: transform 0.2s ease;
}

.lixi.open {
  animation: lixi-open 1.2s ease forwards;
}

@keyframes lixi-open {
  0% {
    transform: scale(1) rotate(0deg);
  }
  15% {
    transform: scale(1.1) rotate(-8deg);
  }
  30% {
    transform: scale(1.1) rotate(8deg);
  }
  45% {
    transform: scale(1.15) rotate(-6deg);
  }
  60% {
    transform: scale(1.2) rotate(0deg);
  }
  100% {
    transform: scale(1.35) translateY(-18px);
    opacity: 0.3;
  }
}

.money-fly {
  position: absolute;
  left: 50%;
  bottom: 180px;
  transform: translateX(-50%);
  pointer-events: none;
}

.money-fly span {
  position: absolute;
  font-size: 28px;
  animation: fly-up 1.6s ease-out forwards;
}

.money-fly span:nth-child(1) { left: -40px; animation-delay: 0s; }
.money-fly span:nth-child(2) { left: -15px; animation-delay: 0.1s; }
.money-fly span:nth-child(3) { left: 10px; animation-delay: 0.15s; }
.money-fly span:nth-child(4) { left: 35px; animation-delay: 0.2s; }
.money-fly span:nth-child(5) { left: 60px; animation-delay: 0.25s; }

@keyframes fly-up {
  0% {
    opacity: 0;
    transform: translateY(0) scale(0.8);
  }
  20% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translateY(-120px) scale(1.2);
  }
}
</style>
