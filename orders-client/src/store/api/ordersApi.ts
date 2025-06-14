import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type { Order, CreateOrderRequest } from '../../types/api'
import { ENV } from '../../config/env'

export const ordersApi = createApi({
  reducerPath: 'ordersApi',
  baseQuery: fetchBaseQuery({
    baseUrl: `${ENV.API_URL}/orders-api`,
  }),
  tagTypes: ['Order'],
  endpoints: (builder) => ({
    createOrder: builder.mutation<Order, CreateOrderRequest>({
      query: (body) => ({
        url: '/orders',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Order'],
    }),
    
    getOrder: builder.query<Order, string>({
      query: (orderId) => `/orders/${orderId}`,
      providesTags: (_result, _error, orderId) => [{ type: 'Order', id: orderId }],
    }),
    
    getUserOrders: builder.query<Order[], string>({
      query: (userId) => `/orders/user/${userId}`,
      transformResponse: (response: Order[] | null) => {
        console.log('API Response for getUserOrders:', response)
        return response || []
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map(({ id }) => ({ type: 'Order' as const, id })),
              { type: 'Order', id: 'LIST' },
            ]
          : [{ type: 'Order', id: 'LIST' }],
    }),
    
    healthCheck: builder.query<{ status: string }, void>({
      query: () => '/info',
    }),
  }),
})

export const {
  useCreateOrderMutation,
  useGetOrderQuery,
  useGetUserOrdersQuery,
  useHealthCheckQuery,
} = ordersApi 