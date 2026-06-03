'use server'
import { cookies } from 'next/headers'

export async function login(formData: FormData) {
	const email = formData.get('email')
	const password = formData.get('password')
	const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

	try {
		const res = await fetch(`${apiUrl}/auth/login`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password }),
		})

		if (!res.ok) {
			return { error: 'Invalid credentials' }
		}

		const data = await res.json()
		const cookieStore = await cookies()
		cookieStore.set('session_jwt', data.token, {
			httpOnly: true,
			secure: process.env.NODE_ENV === 'production',
			sameSite: 'lax',
			path: '/',
			maxAge: 60 * 60 * 24 // 24 hours
		})

		return { success: true }
	} catch (error) {
		return { error: 'Failed to connect to the CRM engine' }
	}
}   
