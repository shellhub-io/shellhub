import Axios from 'axios'
import store from '../store'
import router from "../router";

export default () => {
    let axios = Axios.create({
        baseURL: `${location.protocol}//${location.host}/api`,
        headers: {
            Authorization: `Bearer ${localStorage.getItem('token')}`
        }
    })

    axios.interceptors.response.use(
        response => response,
        (error) => {
            if (error.response.status === 401) {
                store.dispatch('auth/logout')
                router.push('/login')
            }
        }
    )

    return axios
}
