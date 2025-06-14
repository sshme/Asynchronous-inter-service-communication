import { useEffect } from 'react'
import { useSelector, useDispatch } from 'react-redux'
import { Container, VStack, Flex, Box, Heading, Separator, Button } from '@chakra-ui/react'
import { Header } from './components/Header'
import { UserInfo } from './components/UserInfo'
import { OrdersTable } from './components/OrdersTable'
import { CreateOrderButton } from './components/CreateOrderButton'
import { useGetUserOrdersQuery } from './store/api/ordersApi'
import { useCreateAccountMutation } from './store/api/paymentsApi'
import { selectCurrentUserId, selectIsLoggedIn, clearCurrentUser, setCurrentUser } from './store/slices/userSlice'
import type { RootState } from './store'

function App() {
  const dispatch = useDispatch()
  const currentUserId = useSelector((state: RootState) => selectCurrentUserId(state))
  const isLoggedIn = useSelector((state: RootState) => selectIsLoggedIn(state))
  
  const [createAccount, { isLoading: isCreatingAccount }] = useCreateAccountMutation()

  const { 
    data: ordersData, 
    isLoading: isOrdersLoading, 
    error: ordersError,
    isSuccess: isOrdersSuccess 
  } = useGetUserOrdersQuery(currentUserId!, {
    skip: !currentUserId
  })

  const orders = ordersData || []

  console.log('RTK Query Debug:', {
    currentUserId,
    ordersData,
    orders,
    isOrdersLoading,
    ordersError,
    isOrdersSuccess
  })

  useEffect(() => {
    if (!currentUserId && !isLoggedIn && !isCreatingAccount) {
      handleCreateUser()
    }
  }, [currentUserId, isLoggedIn, isCreatingAccount])

  const handleCreateUser = async () => {
    try {
      const accountData = await createAccount().unwrap()
      console.log('Created account:', accountData)
      dispatch(setCurrentUser(accountData.user_id))
    } catch (error) {
      console.error('Failed to create account:', error)
    }
  }

  const handleLogout = () => {
    dispatch(clearCurrentUser())
  }

  if (!currentUserId) {
    return (
      <Container maxW="6xl" py={8}>
        <VStack gap={8} align="center">
          <Heading>
            {isCreatingAccount ? 'Creating your account...' : 'Setting up your session...'}
          </Heading>
        </VStack>
      </Container>
    )
  }

  return (
    <Container maxW="6xl" py={8}>
      <VStack gap={8} align="stretch">
        <Flex align="center" justify="space-between">
          <Header />
          <Button colorPalette="red" variant="outline" onClick={handleLogout}>
            Logout
          </Button>
        </Flex>
        <UserInfo userId={currentUserId} />
        <Separator />
        <Box>
          <Flex align="center" justify="space-between" mb={6}>
            <Heading size="xl">User Orders</Heading>
            <CreateOrderButton userId={currentUserId} />
          </Flex>
          <OrdersTable userId={currentUserId} />
        </Box>
      </VStack>
    </Container>
  )
}

export default App
