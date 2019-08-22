import http from '@/helpers/http'

export const
    login = async (user) => {
        return http().post('/login', user)
    }