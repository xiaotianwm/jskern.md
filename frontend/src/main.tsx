import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import {installDesktopGuards} from './desktopGuards'

const container = document.getElementById('root')

const root = createRoot(container!)

installDesktopGuards()

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
