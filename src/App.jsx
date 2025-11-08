import React, { useState, useEffect } from 'react';

// Tailwind CSS is assumed to be available in the environment.

function App() {
    const [currentMode, setCurrentMode] = useState('Loading...');
    const [message, setMessage] = useState('');
    const [loading, setLoading] = useState(false);

    // Backend API URL now relative, as Caddy will proxy /api/ to the backend
    const API_BASE_URL = '/api'; 

    // Function to fetch the current MPD mode
    const fetchCurrentMode = async () => {
        try {
            setLoading(true);
            const response = await fetch(`${API_BASE_URL}/currentmode`); // Updated endpoint
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            setCurrentMode(data.mode);
            setMessage(`Current MPD Mode: ${data.mode}`);
        } catch (error) {
            console.error('Error fetching current mode:', error);
            setMessage(`Error fetching mode: ${error.message}`);
            setCurrentMode('Unknown');
        } finally {
            setLoading(false);
        }
    };

    // Function to switch MPD mode
    const switchMode = async (mode) => {
        try {
            setLoading(true);
            setMessage(`Switching to ${mode} mode...`);
            // Updated endpoint to include mode in the path
            const response = await fetch(`${API_BASE_URL}/switch/${mode}`, { 
                method: 'GET', // Changed to GET as per new path structure
                headers: {
                    'Content-Type': 'application/json',
                },
                // No body needed for GET request
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(`HTTP error! status: ${response.status} - ${errorData.error || 'Unknown error'}`);
            }

            const data = await response.json();
            setMessage(data.message);
            fetchCurrentMode(); // Refresh status after switching
        } catch (error) {
            console.error('Error switching mode:', error);
            setMessage(`Error switching mode: ${error.message}`);
        } finally {
            setLoading(false);
        }
    };

    // Fetch mode on component mount
    useEffect(() => {
        fetchCurrentMode();
        // Optionally, refresh every few seconds to keep status updated
        const intervalId = setInterval(fetchCurrentMode, 5000); 
        return () => clearInterval(intervalId); // Cleanup on unmount
    }, []);

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 to-gray-800 text-white flex items-center justify-center p-4">
            <div className="bg-gray-800 p-8 rounded-xl shadow-2xl max-w-md w-full border border-gray-700">
                <h1 className="text-4xl font-extrabold text-center mb-6 text-purple-400">
                    MPD Mode Switcher 🎶
                </h1>

                <div className="mb-6 text-center">
                    <p className="text-lg text-gray-300 mb-2">Current MPD Output Mode:</p>
                    <p className={`text-3xl font-bold ${currentMode === 'Exclusive (DSD)' ? 'text-green-400' : currentMode === 'PipeWire' ? 'text-blue-400' : 'text-yellow-400'}`}>
                        {currentMode}
                    </p>
                </div>

                <div className="flex flex-col space-y-4 mb-6">
                    <button
                        onClick={() => switchMode('exclusive')}
                        disabled={loading || currentMode === 'Exclusive (DSD)'}
                        className={`
                            w-full py-3 rounded-lg font-semibold text-lg transition-all duration-300
                            ${currentMode === 'Exclusive (DSD)' ? 'bg-green-700 cursor-not-allowed' : 'bg-green-600 hover:bg-green-500 active:bg-green-700'}
                            ${loading ? 'opacity-50 cursor-not-allowed' : 'shadow-lg hover:shadow-xl'}
                        `}
                    >
                        {loading && currentMode !== 'Exclusive (DSD)' ? 'Switching...' : 'Switch to Exclusive (DSD)'}
                    </button>

                    <button
                        onClick={() => switchMode('pipewire')}
                        disabled={loading || currentMode === 'PipeWire'}
                        className={`
                            w-full py-3 rounded-lg font-semibold text-lg transition-all duration-300
                            ${currentMode === 'PipeWire' ? 'bg-blue-700 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-500 active:bg-blue-700'}
                            ${loading ? 'opacity-50 cursor-not-allowed' : 'shadow-lg hover:shadow-xl'}
                        `}
                    >
                        {loading && currentMode !== 'PipeWire' ? 'Switching...' : 'Switch to PipeWire'}
                    </button>
                </div>

                {message && (
                    <div className="mt-6 p-4 bg-gray-700 rounded-lg text-sm text-gray-200 text-center">
                        {message}
                    </div>
                )}

                <p className="mt-8 text-xs text-gray-500 text-center">
                    Powered by your Gemini coded custom backend and react frontend.
                </p>
            </div>
        </div>
    );
}

export default App;
