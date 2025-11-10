import { useState, useEffect } from 'react'
import { ConfigPart } from './types'

// Tailwind CSS is assumed to be available in the environment.

function App() {
    const [configParts, setConfigParts] = useState<ConfigPart[]>([])
    const [currentMode, setCurrentMode] = useState<ConfigPart>({
        key: '',
        name: 'Loading...',
    })
    const [message, setMessage] = useState('')
    const [loading, setLoading] = useState(false)

    // Backend API URL now relative, as Caddy will proxy /api/ to the backend
    const API_BASE_URL = '/api'

    // Function to fetch the current MPD mode
    const fetchCurrentMode = async () => {
        try {
            setLoading(true)
            const response = await fetch(`${API_BASE_URL}/currentmode`)
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`)
            }
            const data: ConfigPart = await response.json()
            setCurrentMode(data)
        } catch (error: any) {
            console.error('Error fetching current mode:', error)
            setMessage(`Error fetching mode: ${error.message}`)
            setCurrentMode({ key: 'unknown', name: 'Unknown' })
        } finally {
            setLoading(false)
        }
    }

    // Function to fetch available config parts
    const fetchConfigParts = async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/configparts`)
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`)
            }
            const data: ConfigPart[] = await response.json()
            setConfigParts(data)
        } catch (error: any) {
            console.error('Error fetching config parts:', error)
            setMessage(`Error fetching config parts: ${error.message}`)
        }
    }

    // Function to switch MPD mode
    const switchMode = async (modeKey: string, modeName: string) => {
        try {
            setLoading(true)
            setMessage(`Switching to ${modeName} configuration...`)
            const response = await fetch(`${API_BASE_URL}/switch/${modeKey}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            })

            if (!response.ok) {
                const errorData = await response.json()
                throw new Error(
                    `HTTP error! status: ${response.status} - ${
                        errorData.error || 'Unknown error'
                    }`
                )
            }

            const data = await response.json()
            setMessage(data.message)
            fetchCurrentMode() // Refresh status after switching
        } catch (error: any) {
            console.error('Error switching mode:', error)
            setMessage(`Error switching mode: ${error.message}`)
        } finally {
            setLoading(false)
        }
    }

    // Fetch mode and config parts on component mount
    useEffect(() => {
        fetchConfigParts()
        fetchCurrentMode()
        const intervalId = setInterval(fetchCurrentMode, 5000)
        return () => clearInterval(intervalId)
    }, [])

    return (
        <div className='min-h-screen bg-gradient-to-br from-gray-900 to-gray-800 text-white flex items-center justify-center p-4'>
            <div className='bg-gray-800 p-8 rounded-xl shadow-2xl max-w-md w-full border border-gray-700'>
                <h1 className='text-3xl font-extrabold text-center mb-6 text-purple-400'>
                    MPD Mode Switcher 🎶
                </h1>

                <div className='mb-6 text-center'>
                    <p className='text-lg text-gray-300 mb-2'>
                        Current MPD Output Mode:
                    </p>
                    <p className='text-2xl font-bold text-green-400'>
                        {currentMode.name ?? '[Unknown]'}
                    </p>
                </div>

                <div className='flex flex-col space-y-4 mb-6'>
                    {(!configParts || configParts?.length === 0) && (
                        <p className='text-lg text-gray-300 text-center'>
                            You have no mpd configurations to switch/activate.
                        </p>
                    )}
                    {(configParts ?? []).map(part => (
                        <button
                            key={part.key}
                            onClick={() => switchMode(part.key, part.name)}
                            disabled={loading}
                            className={`
                                w-full py-3 rounded-lg font-semibold text-lg transition-all duration-300 cursor-pointer
                                ${
                                    currentMode.key === part.key
                                        ? 'bg-indigo-700'
                                        : 'bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700'
                                }
                                ${
                                    loading && currentMode.key !== part.key
                                        ? 'opacity-50 cursor-not-allowed'
                                        : 'shadow-lg hover:shadow-xl'
                                }
                            `}
                        >
                            {loading && currentMode.key !== part.key
                                ? 'Switching...'
                                : `Switch to ${part.name}`}
                        </button>
                    ))}
                </div>

                {message && (
                    <div className='mt-6 p-4 bg-gray-700 rounded-lg text-sm text-gray-200 text-center'>
                        {message}
                    </div>
                )}
            </div>
        </div>
    )
}

export default App
