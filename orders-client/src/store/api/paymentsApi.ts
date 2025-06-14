import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
import type {
    Account,
    CreateAccountResponse,
    TopUpAccountResponse
} from '../../types/api'
import { ENV } from '../../config/env'

export const paymentsApi = createApi({
    reducerPath: 'paymentsApi',
    baseQuery: fetchBaseQuery({
        baseUrl: `${ENV.API_URL}/payments-api`,
    }),
    tagTypes: ['Account'],
    endpoints: (builder) => ({
        createAccount: builder.mutation<CreateAccountResponse, void>({
            query: () => ({
                url: '/accounts',
                method: 'POST',
            }),
            invalidatesTags: ['Account'],
        }),

        getAccount: builder.query<Account, string>({
            query: (userId) => `/accounts/${userId}`,
            providesTags: (_result, _error, userId) => [{ type: 'Account', id: userId }],
        }),

        topUpAccount: builder.mutation<TopUpAccountResponse, { userId: string; amount: number }>({
            query: ({ userId, amount }) => ({
                url: `/accounts/${userId}/topup`,
                method: 'POST',
                body: { amount },
            }),
            invalidatesTags: (_result, _error, { userId }) => [{ type: 'Account', id: userId }],
        }),

        healthCheck: builder.query<{ status: string }, void>({
            query: () => '/info',
        }),
    }),
})

export const {
    useCreateAccountMutation,
    useGetAccountQuery,
    useTopUpAccountMutation,
    useHealthCheckQuery: usePaymentsHealthCheckQuery,
} = paymentsApi 