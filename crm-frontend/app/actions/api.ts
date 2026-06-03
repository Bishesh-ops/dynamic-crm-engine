'use server'
import { cookies } from "next/headers"

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://127.0.0.1:8080'

export async function fetchFromAPI(endpoint: string, options: RequestInit = {}) {
	const cookieStore = await cookies()
	const token = cookieStore.get('session_jwt')?.value

	if (!token) {
		throw new Error('Unauthorized: No session token found')
	}

	const res = await fetch(`${API_URL}/${endpoint.replace(/^\//, '')}`, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			'Authorization': `Bearer ${token}`,
			...options.headers,
		},
	})

	if (!res.ok) {
		throw new Error(`API Error: ${res.statusText}`)
	}
	return res.json()
}

export async function getSchema(schemaName: string) {
	return fetchFromAPI(`api/schemas/${schemaName}`)
}   
