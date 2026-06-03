'use client'
import { useState } from 'react'
import { fetchFromAPI } from '@/app/actions/api'

type FieldDef = {
  type: string
  required: boolean
  target_schema?: string
}

type Schema = {
  name: string
  fields: Record<string, FieldDef>
}

export default function DynamicForm({ schema }: { schema: Schema }) {
  const [formData, setFormData] = useState<Record<string, any>>({})

  const handleChange = (key: string, value: any) => {
    setFormData((prev) => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault()
    try{
      await fetchFromAPI(`/api/entities/${schema.name}`,{
        method: 'POST',
        body: JSON.stringify(formData)
      })
      alert(`Sucess! New ${schema.name} record queued for processing.`)
      setFormData({})
      e.target.reset()
    }catch(error: any){
      alert(`Failed to save record: ${error.essage}.`)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6 bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <h2 className="text-xl font-bold text-gray-900 capitalize">
        Create New {schema.name.replace('_', ' ')}
      </h2>

      <div className="space-y-4">
        {Object.entries(schema.fields).map(([fieldName, def]) => {
          return (
            <div key={fieldName}>
              <label className="block text-sm font-semibold text-gray-700 capitalize mb-1">
                {fieldName.replace('_', ' ')} {def.required && <span className="text-red-500">*</span>}
              </label>

              {def.type === 'string' && (
                <input
                  type="text"
                  required={def.required}
                  onChange={(e) => handleChange(fieldName, e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                />
              )}

              {def.type === 'int' && (
                <input
                  type="number"
                  required={def.required}
                  onChange={(e) => handleChange(fieldName, parseInt(e.target.value))}
                  className="w-full p-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
                />
              )}

              {def.type === 'relation' && (
                <select
                  required={def.required}
                  onChange={(e) => handleChange(fieldName, e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-md bg-gray-50 text-gray-500"
                >
                  <option value="">Select {def.target_schema}...</option>
                  <option value="fake-uuid-123">John Doe (Simulated UUID)</option>
                </select>
              )}
            </div>
          )
        })}
      </div>

      <button
        type="submit"
        className="w-full bg-blue-600 text-white p-2 rounded-md hover:bg-blue-700 transition font-medium"
      >
        Save Record
      </button>
    </form>
  )
}   
