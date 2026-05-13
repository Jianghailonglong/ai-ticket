import api from './index'

export const ticketApi = {
  list(params?: { status?: string; page?: number; page_size?: number; filter?: string }) {
    return api.get('/tickets', { params })
  },
  get(id: string) {
    return api.get(`/tickets/${id}`)
  },
  create(data: {
    title: string
    scene: string
    approver: string
    reason?: string
  }) {
    return api.post('/tickets', data)
  },
  approve(id: string, comment?: string) {
    return api.post(`/tickets/${id}/approve`, { comment })
  },
  reject(id: string, comment: string) {
    return api.post(`/tickets/${id}/reject`, { comment })
  }
}
