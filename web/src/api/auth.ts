import api from './index'

export const authApi = {
  register(data: { username: string; password: string; display_name?: string }) {
    return api.post('/auth/register', data)
  },
  login(data: { username: string; password: string }) {
    return api.post('/auth/login', data)
  }
}
