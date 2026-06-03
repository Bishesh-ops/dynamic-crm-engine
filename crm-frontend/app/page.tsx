'use client'

import { useState } from 'react'
import { login } from './actions/auth'
import { useRouter } from 'next/navigation'

export default function LoginPage() {
  const [error, setError] = useState('')
  const router = useRouter()

  async function handleSubmit(formData: FormData) {
    const result = await login(formData)
    
    if (result?.error) {
      setError(result.error)
    } else {
      router.push('/dashboard') 
    }
  }

  return (
    <div className="flex h-screen items-center justify-center bg-purple-50">
      <div className="w-full max-w-md p-8 bg-white rounded-lg shadow-md border border-gray-200">
        <h1 className="text-2xl font-bold mb-6 text-gray-900">CRM Engine Login</h1>
        
        {error && <p className="text-red-500 mb-4 text-sm">{error}</p>}

        <form action={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-900">Email</label>
            <input 
              type="email" 
              name="email" 
              required 
              className="mt-1 block w-full p-2 border text-gray-900 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500" 
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-900">Password</label>
            <input 
              type="password" 
              name="password" 
              required 
              className="mt-1 block w-full p-2 border text-gray-900 border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500" 
            />
          </div>
          <button 
            type="submit" 
            className="w-full bg-gray-900 text-white p-2 rounded-md hover:bg-gray-800 transition"
          >
            Sign In
          </button>
        </form>
      </div>
    </div>
  )
}
