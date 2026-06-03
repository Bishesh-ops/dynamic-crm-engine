import { getSchema } from '../actions/api'
import DynamicForm from '../../components/DynamicForm'

export default async function DashboardPage() {
  let schema = null
  let error = null

  try {
    schema = await getSchema('enterprise_leads')
  } catch (e: any) {
    error = e.message
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        <header className="mb-8 flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900">Workspace Dashboard</h1>
          <div className="px-4 py-2 bg-gray-200 text-gray-700 rounded-full text-sm font-medium">
            Admin User
          </div>
        </header>

        {error ? (
          <div className="p-4 bg-red-100 text-red-700 rounded-md border border-red-200">
            Failed to load schema: {error}. Did you run the load_test.sh to create the schema?
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Left Column: The Form */}
            <div className="text-gray-900">
              {schema && <DynamicForm schema={schema} />}
            </div>

            {/* Right Column: Placeholder for the Data Grid */}
            <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200 flex items-center justify-center text-gray-400">
              Data grid view coming soon...
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
