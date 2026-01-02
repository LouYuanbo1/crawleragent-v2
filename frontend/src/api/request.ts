import axios from 'axios'

interface RequestConfig {
  url: string;
  method: string;
  data?: any;
  params?: Record<string, any>;
}

async function request(config: RequestConfig) {
  try {
    const resp = await axios({
      url: config.url,
      method: config.method,
      data: config.data,
      params: config.params,
    })
    console.log('resp:', resp)
    return resp.data
  } catch (err) {
    return Promise.reject(err)
  }
}

export default request
