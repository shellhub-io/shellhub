import http from '@/helpers/http'

export const
    fetchSessions = async () => {
        return http().get('/sessions')
    },
    getSession = async (uid) => {
        return http().get(`/session/${uid}`)
    }

