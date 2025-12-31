<script setup>
import { ref } from 'vue'

const id = ref('')
const account = ref('')
const bank = ref('')
const bankno = ref('')
const fullname = ref('')
const error = ref('')
const loading = ref(false)

const isOpening = ref(false)
const showMoneyFly = ref(false)
const showResultText = ref(false)
const amount = ref(0)
const opened = ref(false)

const fireCanvas = ref(null)

async function submit() {
  if (opened.value) return

  error.value = ''
  showResultText.value = false
  showMoneyFly.value = false

  if (!id.value || !account.value || !bank.value || !bankno.value || !fullname.value) {
    error.value = 'Vui lòng nhập đầy đủ thông tin trước khi mở lì xì'
    return
  }

  loading.value = true
  isOpening.value = true
  opened.value = true

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

    amount.value = data.amount

    setTimeout(() => {
      showMoneyFly.value = true
      showResultText.value = true
      launchFireworks()
    }, 3000)

  } catch (e) {
    error.value = e.message || 'Có lỗi xảy ra'
    opened.value = false
  } finally {
    setTimeout(() => {
      isOpening.value = false
      loading.value = false
    }, 3000)
  }
}

function launchFireworks(duration = 6000) {
  const canvas = fireCanvas.value
  const ctx = canvas.getContext('2d')

  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  let particles = []

  function createFirework() {
    const x = Math.random() * canvas.width
    const y = Math.random() * canvas.height * 0.5
    const colors = ['#ff0000', '#ffd700', '#ff8c00', '#ffffff']

    for (let i = 0; i < 36; i++) {
      particles.push({
        x,
        y,
        vx: (Math.random() - 0.5) * 6,
        vy: (Math.random() - 0.5) * 6,
        alpha: 1,
        color: colors[Math.floor(Math.random() * colors.length)],
      })
    }
  }

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    particles.forEach(p => {
      ctx.globalAlpha = p.alpha
      ctx.fillStyle = p.color
      ctx.beginPath()
      ctx.arc(p.x, p.y, 3, 0, Math.PI * 2)
      ctx.fill()

      p.x += p.vx
      p.y += p.vy
      p.alpha -= 0.02
    })
    particles = particles.filter(p => p.alpha > 0)
    requestAnimationFrame(draw)
  }

  draw()
  const interval = setInterval(createFirework, 320)

  setTimeout(() => {
    clearInterval(interval)
    particles = []
    ctx.clearRect(0, 0, canvas.width, canvas.height)
  }, duration)
}
</script>

<template>
  <div class="user-page">
    <canvas ref="fireCanvas" class="firework-canvas"></canvas>

    <div class="card text-center">

      <h1 class="text-3xl font-bold text-red-700 mb-3">🧧 Lì xì 2026</h1>

      <input v-model="id" :disabled="opened" placeholder="Mã lì xì" class="input" />
      <input v-model="account" :disabled="opened" placeholder="Tên ingame" class="input" />
      <input v-model="bank" :disabled="opened" placeholder="Tên ngân hàng" class="input" />
      <input v-model="bankno" :disabled="opened" placeholder="Số tài khoản" class="input" />
      <input v-model="fullname" :disabled="opened" placeholder="Họ và tên" class="input mb-4" />

      <button
        @click="submit"
        :disabled="loading || opened"
        class="btn"
      >
        <span v-if="!opened">
          {{ loading ? 'Đang mở bao lì xì...' : 'Mở lì xì 🧧' }}
        </span>
        <span v-else>
          Đã mở lì xì 🧧
        </span>
      </button>

      <div v-if="showResultText" class="result">
        <p class="result-money">🎉 Bạn nhận được {{ amount }}.000₫ 🎉</p>
        <p class="result-text">
          ✨ <b>Kết sổ 2025 – Say Hi 2026</b><br />
          BQT FAM chúc bạn một năm Ngọ chạy thật nhanh để gặt nhiều thành công,
          bền bỉ và thắng lớn trên đường đua của riêng mình 🏇🔥
        </p>
      </div>

      <p v-if="error" class="text-red-600 mt-3">{{ error }}</p>
      <div class="lixi-row">
        <div
          class="lixi"
          :class="{ open: isOpening }"
          v-for="i in 4"
          :key="i"
        >
          🧧
        </div>
      </div>

      <div v-if="showMoneyFly" class="money-fly">
        <span v-for="i in 7" :key="i">💸</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-page {
  min-height: 100vh;
  background: url('/4261727762388106724.jpg') center/cover no-repeat;
  display: flex;
  justify-content: center;
  align-items: center;
}

.card {
  width: 360px;
  padding: 28px;
  background: rgba(255,248,220,.96);
  border-radius: 20px;
  box-shadow: 0 10px 30px rgba(0,0,0,.25);
  position: relative;
}

.input {
  width: 100%;
  margin-bottom: 10px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #ddd;
}

.input:disabled {
  background: #f5f5f5;
  opacity: .7;
}

.btn {
  width: 100%;
  padding: 12px;
  font-weight: bold;
  background: #dc2626;
  color: white;
  border-radius: 12px;
}

.btn:disabled {
  opacity: .6;
  cursor: not-allowed;
}

.lixi-row {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 20px;
}

.lixi {
  font-size: 64px;
}

.lixi.open {
  animation: lixi-open 1.2s ease forwards;
}

@keyframes lixi-open {
  0% { transform: scale(1); }
  100% { transform: scale(1.25) translateY(-16px); opacity: .3; }
}

.money-fly {
  position: absolute;
  left: 50%;
  bottom: 170px;
  transform: translateX(-50%);
  pointer-events: none;
  z-index: 10;
}

.money-fly span {
  position: absolute;
  font-size: 26px;
  opacity: 0;
  animation: fly-up 1.6s ease forwards;
}

.money-fly span:nth-child(1) { left: -36px; animation-delay: 0s; }
.money-fly span:nth-child(2) { left: -24px; animation-delay: .08s; }
.money-fly span:nth-child(3) { left: -12px; animation-delay: .16s; }
.money-fly span:nth-child(4) { left: 0px; animation-delay: .24s; }
.money-fly span:nth-child(5) { left: 12px; animation-delay: .32s; }
.money-fly span:nth-child(6) { left: 24px; animation-delay: .40s; }
.money-fly span:nth-child(7) { left: 36px; animation-delay: .48s; }

@keyframes fly-up {
  0% { opacity: 0; transform: translateY(0) scale(.8); }
  20% { opacity: 1; }
  100% { opacity: 0; transform: translateY(-140px) scale(1.15); }
}

.result {
  margin-top: 16px;
  animation: fade-in .6s ease;
}

.result-money {
  font-size: 16px;
  font-weight: 600;
  color: #15803d;
}

.result-text {
  font-size: 13px;
  margin-top: 6px;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.firework-canvas {
  position: fixed;
  inset: 0;
  pointer-events: none;
}
</style>
