export const ENV = {
    API_URL: import.meta.env.VITE_API_URL || 'http://localhost:80',
} as const

const validateEnv = () => {
    console.log('Environment configuration:', ENV)
    console.log('Orders API URL:', `${ENV.API_URL}/orders-api`)
    console.log('Payments API URL:', `${ENV.API_URL}/payments-api`)
    
    if (!import.meta.env.VITE_API_URL) {
        console.warn('VITE_API_URL not set, using default:', ENV.API_URL)
    }
}

if (import.meta.env.DEV) {
    validateEnv()
} 