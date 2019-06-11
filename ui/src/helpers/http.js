import Axios from 'axios'

export default Axios.create({
    baseURL: `http://${location.host}/api`,
})
