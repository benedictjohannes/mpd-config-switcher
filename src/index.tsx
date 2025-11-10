import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './App.css'

const rootDomNode = document.getElementById('root')
if (rootDomNode) {
    const root = ReactDOM.createRoot(rootDomNode)
    root.render(
        <React.StrictMode>
            <App />
        </React.StrictMode>
    )
}
