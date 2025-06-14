import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'

interface UserState {
    currentUserId: string | null
    isLoggedIn: boolean
}

const loadUserFromStorage = (): UserState => {
    try {
        const savedUser = localStorage.getItem('appMarket_user')
        if (savedUser) {
            return JSON.parse(savedUser)
        }
    } catch (error) {
        console.error('Error loading user from localStorage:', error)
    }

    return {
        currentUserId: null,
        isLoggedIn: false
    }
}

const saveUserToStorage = (state: UserState) => {
    try {
        localStorage.setItem('appMarket_user', JSON.stringify(state))
    } catch (error) {
        console.error('Error saving user to localStorage:', error)
    }
}

const initialState: UserState = loadUserFromStorage()

const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {
        setCurrentUser: (state, action: PayloadAction<string>) => {
            state.currentUserId = action.payload
            state.isLoggedIn = true
            saveUserToStorage(state)
        },
        clearCurrentUser: (state) => {
            state.currentUserId = null
            state.isLoggedIn = false
            saveUserToStorage(state)
        },
    }
})

export const { setCurrentUser, clearCurrentUser } = userSlice.actions
export default userSlice.reducer

export const selectCurrentUserId = (state: { user: UserState }) => state.user.currentUserId
export const selectIsLoggedIn = (state: { user: UserState }) => state.user.isLoggedIn 