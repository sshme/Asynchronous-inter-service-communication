import { Button } from '@chakra-ui/react'
import { FaPlus } from 'react-icons/fa'
import { useCreateOrderMutation } from '../store/api/ordersApi'

interface CreateOrderButtonProps {
    userId: string
}

export const CreateOrderButton = ({ userId }: CreateOrderButtonProps) => {
    const [createOrder, { isLoading }] = useCreateOrderMutation()

    const handleCreateOrder = async () => {
        try {
            await createOrder({ user_id: userId }).unwrap()
        } catch (error) {
            console.error('Failed to create order:', error)
        }
    }

    return (
        <Button
            colorPalette="purple"
            onClick={handleCreateOrder}
            loading={isLoading}
        >
            <FaPlus />
            Create Order
        </Button>
    )
} 