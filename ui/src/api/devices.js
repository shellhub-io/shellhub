import http from '@/helpers/http'

export const
    fetchDevices = async () => {
        return http().get('/devices')
    }