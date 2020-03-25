import http from '@/helpers/http'

export const
    fetchSessions = async (per_page, page) => {
        return http().get('/sessions?per_page='+per_page.toString()+'&page='+page.toString())
    },
    getSession = async (uid) => {
        return http().get(`/sessions/${uid}`)
    },
    closeSession = async (session) => {
        return http().post(`/sessions/${session.uid}/close`, { device: session.device_uid })
    }

