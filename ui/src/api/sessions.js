import http from '@/helpers/http'

export const
    fetchSessions = async () => {
        return http.get('/sessions')
    }