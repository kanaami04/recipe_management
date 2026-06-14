import axios from 'axios';
import type { AxiosInstance } from 'axios';
import type { Token } from '@/type/LoginUser';

export const api: AxiosInstance = axios.create({
	baseURL: 'http://127.0.0.1:8000',
	withCredentials: false,  // JWTは手動で送る
});

export const fetcher = (url: string, token: Token) => api.get(url, {
		headers: { Authorization: `Bearer ${token}` }
	}).then(res => res.data)