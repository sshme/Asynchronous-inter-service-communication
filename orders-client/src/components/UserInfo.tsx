import { Box, VStack, HStack, Heading, Text, Badge, Button, Spinner } from '@chakra-ui/react'
import { FaMoneyBillWave } from 'react-icons/fa'
import { useGetAccountQuery, useTopUpAccountMutation } from '../store/api/paymentsApi'

interface UserInfoProps {
    userId: string
}

export const UserInfo = ({ userId }: UserInfoProps) => {
    const { data: account, isLoading: isAccountLoading, error } = useGetAccountQuery(userId)
    const [topUpAccount, { isLoading: isTopingUp }] = useTopUpAccountMutation()

    const handleTopupAccount = async () => {
        if (!account) return
        
        try {
            await topUpAccount({ 
                userId: account.user_id, 
                amount: 100
            }).unwrap()
        } catch (error) {
            console.error('Failed to topup account:', error)
        }
    }

    return (
        <Box p={6} borderWidth="1px" borderRadius="lg" bg="gray.50" _dark={{ bg: "gray.800" }}>
            <VStack align="start" gap={4}>
                <Heading size="lg">Current User</Heading>
                
                <HStack gap={4}>
                    <Text fontSize="lg" fontWeight="semibold">
                        User ID: 
                    </Text>
                    <Badge colorPalette="blue" size="lg" px={3} py={1}>
                        {userId}
                    </Badge>
                </HStack>

                {isAccountLoading && (
                    <HStack gap={2}>
                        <Spinner size="sm" />
                        <Text>Loading account info...</Text>
                    </HStack>
                )}

                {error && (
                    <Text color="red.500" fontSize="sm">
                        Failed to load account information
                    </Text>
                )}

                {account && !isAccountLoading && (
                    <HStack gap={4}>
                        <Text fontSize="lg" fontWeight="semibold">
                            Balance: 
                        </Text>
                        <Badge colorPalette="green" size="lg" px={3} py={1}>
                            ${account.balance.toFixed(2)}
                        </Badge>
                    </HStack>
                )}

                {account && (
                    <HStack gap={4}>
                        <Button 
                            colorPalette="green" 
                            variant="outline"
                            onClick={handleTopupAccount}
                            loading={isTopingUp}
                        >
                            <FaMoneyBillWave />
                            Topup $100
                        </Button>
                    </HStack>
                )}
            </VStack>
        </Box>
    )
} 