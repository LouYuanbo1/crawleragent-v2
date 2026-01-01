import axios from 'axios'


async function request(url: string, method: string, data?: any) {
  try {
    const resp = await axios({
      url,
      method,
      data,
    })
    return resp.data
  } catch (err) {
    return Promise.reject(err)
  }
}

export default request
