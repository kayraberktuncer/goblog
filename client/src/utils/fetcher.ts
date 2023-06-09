import axios, { AxiosRequestConfig } from 'axios'

export const axiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  withCredentials: true,
})

axiosInstance.interceptors.response.use(
  function (response) {
    return response
  },
  function (error) {
    return Promise.reject(error.response.data)
  }
)

export const fetcher = async (url: any, config?: AxiosRequestConfig) => {
  const res = await axiosInstance.get(url, config)
  return res.data
}
