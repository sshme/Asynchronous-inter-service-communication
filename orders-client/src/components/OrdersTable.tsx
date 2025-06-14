import { useEffect } from 'react'
import { useDispatch } from 'react-redux'
import { Box, Text, Badge } from '@chakra-ui/react'
import { useGetUserOrdersQuery, ordersApi } from '../store/api/ordersApi'
import { paymentsApi } from '../store/api/paymentsApi'
import type { Order, OrderStatus } from '../types/api'
import { ENV } from '../config/env'

interface OrdersTableProps {
    userId: string
}

export const OrdersTable = ({ userId }: OrdersTableProps) => {
    const dispatch = useDispatch()
    const { data: orders = [], isLoading } = useGetUserOrdersQuery(userId)

    useEffect(() => {
        const eventSource = new EventSource(`${ENV.API_URL}/orders-api/orders/stream?user_id=${userId}`)

        eventSource.addEventListener('connected', (event) => {
            console.log('Connected to SSE:', event.data)
        })

        eventSource.addEventListener('order-update', (event) => {
            try {
                const updatedOrder: Order = JSON.parse(event.data)
                console.log('updatedOrder', updatedOrder)

                dispatch(ordersApi.util.invalidateTags([
                    { type: 'Order', id: updatedOrder.id },
                    { type: 'Order', id: 'LIST' }
                ]))

                dispatch(paymentsApi.util.invalidateTags([
                    { type: 'Account', id: updatedOrder.userID }
                ]))
            } catch (error) {
                console.error('Error parsing SSE data:', error)
            }
        })

        eventSource.onmessage = (event) => {
            console.log('General SSE message:', event)
        }

        eventSource.onerror = (error) => {
            console.error('SSE connection error:', error)
            eventSource.close()
        }

        return () => {
            eventSource.close()
        }
    }, [userId, dispatch])

    const getStatusColor = (status: OrderStatus) => {
        switch (status) {
            case 'completed': return 'green'
            case 'paid': return 'blue'
            case 'payment_pending': return 'yellow'
            case 'payment_failed': return 'red'
            case 'created': return 'gray'
            case 'cancelled': return 'red'
            default: return 'gray'
        }
    }

    if (isLoading) {
        return (
            <Box p={4} textAlign="center">
                <Text>Loading orders...</Text>
            </Box>
        )
    }

    return (
        <Box borderWidth="1px" borderRadius="lg" bg="white" _dark={{ bg: "gray.800" }} overflow="hidden">
            <Box overflowX="auto">
                <Box as="table" w="full">
                    <Box as="thead" bg="gray.50" _dark={{ bg: "gray.700" }}>
                        <Box as="tr">
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Order ID</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">User ID</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Status</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Amount</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Payment ID</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Error Reason</Box>
                            <Box as="th" p={4} textAlign="left" fontWeight="semibold">Updated At</Box>
                        </Box>
                    </Box>
                    <Box as="tbody">
                        {orders.map((order) => (
                            <Box as="tr" key={order.id} borderTopWidth="1px">
                                <Box as="td" p={4} fontFamily="mono" fontWeight="medium">
                                    {order.id}
                                </Box>
                                <Box as="td" p={4} fontFamily="mono">
                                    {order.userID}
                                </Box>
                                <Box as="td" p={4}>
                                    <Badge
                                        colorPalette={getStatusColor(order.status)}
                                        textTransform="capitalize"
                                    >
                                        {order.status.replace('_', ' ')}
                                    </Badge>
                                </Box>
                                <Box as="td" p={4} fontFamily="mono">
                                    ${order.amount?.toFixed(2) || '0.00'} {order.currency || 'USD'}
                                </Box>
                                <Box as="td" p={4} fontFamily="mono">
                                    {order.paymentID || '-'}
                                </Box>
                                <Box as="td" p={4}>
                                    {order.errorReason ? (
                                        <Text color="red.500" fontSize="sm">
                                            {order.errorReason}
                                        </Text>
                                    ) : (
                                        <Text color="gray.500">-</Text>
                                    )}
                                </Box>
                                <Box as="td" p={4} fontFamily="mono" fontSize="sm">
                                    {new Date(order.updatedAt).toLocaleString()}
                                </Box>
                            </Box>
                        ))}
                        {orders.length === 0 && (
                            <Box as="tr">
                                <Box as="td" p={4} textAlign="center" color="gray.500" style={{ gridColumn: '1 / -1' }}>
                                    No orders found
                                </Box>
                            </Box>
                        )}
                    </Box>
                </Box>
            </Box>
        </Box>
    )
} 