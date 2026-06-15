import useSWR from 'swr'

import { fetcher } from '@/lib/api'
import type { Token } from '@/type/LoginUser'

export function useApiData<T>(url: string, token: Token) {
  const shouldFetch = !!url && !!token
  const { data, error } = useSWR<T>(shouldFetch ? url : null, (url) => fetcher(url, token), {
    revalidateOnFocus: false,
    refreshInterval: 0,
  })
  return { data, error }
}
