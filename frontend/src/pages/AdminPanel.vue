<script setup>
import { ref, onMounted, computed } from 'vue'
import {
  adminLogin,
  adminLogout,
  fetchCurrentPool,
  createPool,
  fetchClaimHistory,
  fetchUserIDs,
  addUserIDs,
  deleteUserID,
  setSession,
} from '../api'

const loggedIn = ref(!!localStorage.getItem('admin_session'))
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const activeTab = ref('ids')

const userIDs = ref([])
const claimHistory = ref([])

const DENOMS = [50, 100, 200, 500]
const currentPool = ref(DENOMS.map(a => ({ amount: a, qty: 0 })))
const editPool = ref(DENOMS.map(a => ({ amount: a, qty: 0 })))

const drawnSet = computed(() => {
  const s = new Set()
  claimHistory.value.forEach(c => {
    if (c.id) s.add(c.id)
  })
  return s
})

function randomHex16() {
  return [...crypto.getRandomValues(new Uint8Array(8))]
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
}

async function login() {
  error.value = ''
  loading.value = true
  try {
    const res = await adminLogin(username.value, password.value)
    setSession(res.session)
    loggedIn.value = true
    await loadAll()
  } catch {
    error.value = 'Sai tài khoản hoặc mật khẩu'
  } finally {
    loading.value = false
  }
}

async function logout() {
  try { await adminLogout() } catch {}
  localStorage.removeItem('admin_session')
  loggedIn.value = false
  userIDs.value = []
  claimHistory.value = []
}

async function loadAll() {
  await Promise.all([
    loadUserIDs(),
    loadPool(),
    loadClaims(),
  ])
}

async function loadUserIDs() {
  try {
    const res = await fetchUserIDs()
    userIDs.value = res.items || []
  } catch {
    userIDs.value = []
  }
}

async function loadPool() {
  try {
    const res = await fetchCurrentPool()
    const items = res.items || []
    currentPool.value = DENOMS.map(a => {
      const f = items.find(i => i.amount === a)
      return { amount: a, qty: f ? f.qty : 0 }
    })
    editPool.value = currentPool.value.map(i => ({ ...i }))
  } catch {}
}

async function loadClaims() {
  try {
    const res = await fetchClaimHistory()
    claimHistory.value = res.items || []
  } catch {
    claimHistory.value = []
  }
}

async function addRandomID() {
  const id = randomHex16()
  try {
    await addUserIDs([id])
    userIDs.value.push(id)
  } catch {
    alert('Không thêm được mã')
  }
}

async function removeID(id) {
  if (drawnSet.value.has(id)) {
    alert('Mã này đã rút – không thể xoá')
    return
  }
  if (!confirm(`Xoá mã ${id}?`)) return
  try {
    await deleteUserID(id)
    userIDs.value = userIDs.value.filter(i => i !== id)
  } catch {
    alert('Không xoá được mã')
  }
}

async function savePool() {
  loading.value = true
  try {
    await createPool(editPool.value)
    currentPool.value = editPool.value.map(i => ({ ...i }))
    alert('Đã cập nhật danh sách lì xì hiện tại')
  } catch {
    alert('Lỗi cập nhật lì xì')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (loggedIn.value) loadAll()
})
</script>

<template>
  <div class="admin-page">
  <div class="min-h-screen flex items-center justify-center px-4">

    <div v-if="!loggedIn" class="bg-yellow-50 p-6 rounded-xl w-full max-w-sm">
      <h1 class="text-xl font-bold text-center mb-4">Admin Panel</h1>
      <input v-model="username" class="border p-2 w-full mb-3" placeholder="Username" />
      <input v-model="password" type="password" class="border p-2 w-full mb-3" placeholder="Password" />
      <button @click="login" class="bg-red-600 text-white w-full py-2 rounded">
        Đăng nhập
      </button>
      <p v-if="error" class="text-center text-sm text-red-600 mt-2">{{ error }}</p>
    </div>

    <div
      v-else
      class="
        bg-yellow-50
        p-4 sm:p-5 md:p-6
        rounded-xl
        shadow-xl
        space-y-5
        w-full
        max-w-[860px]
        mx-auto
      "
    >

      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">Admin Panel</h1>
        <button @click="logout" class="text-sm underline text-red-700">Đăng xuất</button>
      </div>

      <div class="flex gap-2 border-b pb-2">
        <button @click="activeTab='ids'" :class="activeTab==='ids' ? 'bg-red-600 text-white' : 'bg-gray-200'"
          class="px-4 py-1 rounded-t text-sm">Mã rút</button>
        <button @click="activeTab='pool'" :class="activeTab==='pool' ? 'bg-red-600 text-white' : 'bg-gray-200'"
          class="px-4 py-1 rounded-t text-sm">Lì xì</button>
        <button @click="activeTab='claims'" :class="activeTab==='claims' ? 'bg-red-600 text-white' : 'bg-gray-200'"
          class="px-4 py-1 rounded-t text-sm">Danh sách đã nhận</button>
      </div>

      <div v-if="activeTab==='ids'" class="bg-white p-4 rounded border">
        <div class="flex justify-between mb-3">
          <h2 class="font-semibold">Danh sách mã</h2>
          <button @click="addRandomID" class="bg-blue-600 text-white px-3 py-1 rounded text-sm">
            + Thêm mã
          </button>
        </div>

        <table class="w-full text-sm border">
          <thead class="bg-gray-100">
            <tr>
              <th class="border px-2 py-1">Mã</th>
              <th class="border px-2 py-1 w-32">Trạng thái</th>
              <th class="border px-2 py-1 w-16">Xoá</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="id in userIDs" :key="id">
              <td class="border px-2 py-1 font-mono text-xs">{{ id }}</td>
              <td class="border px-2 py-1 text-center">
                <span
                  v-if="drawnSet.has(id)"
                  class="text-red-600 font-semibold"
                >
                  Đã kích hoạt
                </span>
                <span
                  v-else
                  class="text-green-600 font-semibold"
                >
                  Chưa kích hoạt
                </span>
              </td>
              <td class="border text-center">
                <button
                  @click="removeID(id)"
                  :class="drawnSet.has(id) ? 'opacity-30 cursor-not-allowed' : 'text-red-600'"
                >
                  ❌
                </button>
              </td>
            </tr>

            <tr v-if="userIDs.length === 0">
              <td colspan="3" class="text-center py-4 text-gray-400">Chưa có mã</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="activeTab==='pool'" class="grid grid-cols-2 gap-4">
        <div>
          <h2 class="font-semibold mb-2">Danh sách lì xì hiện tại</h2>
          <table class="w-full text-sm border">
            <tr v-for="p in currentPool" :key="p.amount">
              <td class="border px-2 py-1">{{ p.amount }}.000₫</td>
              <td class="border text-center">{{ p.qty }}</td>
            </tr>
          </table>
        </div>

        <div>
          <h2 class="font-semibold mb-2">Cập nhật lại danh sách lì xì hiện tại</h2>
          <table class="w-full text-sm border">
            <tr v-for="p in editPool" :key="p.amount">
              <td class="border px-2 py-1">{{ p.amount }}.000₫</td>
              <td class="border text-center">
                <input type="number" min="0" v-model.number="p.qty"
                  class="border w-16 text-center" />
              </td>
            </tr>
          </table>
          <button @click="savePool" class="mt-3 bg-red-600 text-white w-full py-2 rounded">
            Cập nhật danh sách lì xì
          </button>
        </div>
      </div>

      <div v-if="activeTab==='claims'" class="bg-white p-4 rounded border">
        <h2 class="font-semibold mb-2">Lịch sử rút</h2>
        <table class="w-full text-sm border">
          <thead class="bg-gray-100">
            <tr>
              <th class="border px-2 py-1">Mã</th>
              <th class="border px-2 py-1">Tên ingame</th>
              <th class="border px-2 py-1">Tên ngân hàng</th>
              <th class="border px-2 py-1">Số tài khoản</th>
              <th class="border px-2 py-1">Tiền</th>
              <th class="border px-2 py-1">Họ và tên</th>
              <th class="border px-2 py-1">Thời gian</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in claimHistory" :key="c.time">
              <td class="border px-2 py-1 font-mono text-xs">{{ c.id }}</td>
              <td class="border px-2 py-1">{{ c.account }}</td>
              <td class="border px-2 py-1">{{ c.bank }}</td>
              <td class="border px-2 py-1">{{ c.bankno }}</td>
              <td class="border px-2 py-1 text-red-600 font-bold">
                {{ c.amount }}.000₫
              </td>
              <td class="border px-2 py-1">{{ c.name }}</td>
              <td class="border px-2 py-1 text-xs">
                {{ new Date(c.time).toLocaleString() }}
              </td>
            </tr>
            <tr v-if="claimHistory.length === 0">
              <td colspan="5" class="text-center py-4 text-gray-400">
                Chưa có ai rút
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>
  </div>
  </div>
</template>
<style scoped>
.admin-page {
  min-height: 100vh;
  background-image: url('/4261727762388106724.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
