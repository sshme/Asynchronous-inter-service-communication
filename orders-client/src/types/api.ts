export type OrderStatus = 
  | 'created' 
  | 'payment_pending' 
  | 'paid' 
  | 'payment_failed' 
  | 'completed' 
  | 'cancelled'

export interface Order {
  id: string
  userID: string
  amount: number
  currency: string
  status: OrderStatus
  paymentID: string
  errorReason: string
  createdAt: string
  updatedAt: string
}

export interface CreateOrderRequest {
  user_id: string
}

export interface Account {
  id: string
  user_id: string
  balance: number
  created_at: string
  updated_at: string
}

export interface CreateAccountResponse {
  id: string
  user_id: string
  balance: number
  created_at: string
}

export interface TopUpAccountRequest {
  amount: number
}

export interface TopUpAccountResponse {
  id: string
  user_id: string
  balance: number
  updated_at: string
}

export interface ErrorResponse {
  error: string
} 