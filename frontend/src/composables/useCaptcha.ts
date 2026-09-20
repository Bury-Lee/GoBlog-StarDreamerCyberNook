import { reactive, ref } from 'vue'
import { fetchCaptcha } from '@/api/user'

export function useCaptcha(target: string) {
  const state = reactive({
    captchaID: '',
    captchaCode: '',
    image: '',
  })
  const loading = ref(false)

  async function load(): Promise<void> {
    loading.value = true
    try {
      const data = await fetchCaptcha(target)
      state.captchaID = data?.captchaID || ''
      state.image = data?.captcha || ''
      state.captchaCode = ''
    } catch {
      state.captchaID = ''
      state.image = ''
    } finally {
      loading.value = false
    }
  }

  function reset(): void {
    state.captchaID = ''
    state.captchaCode = ''
    state.image = ''
  }

  return reactive({ state, loading, load, reset })
}
