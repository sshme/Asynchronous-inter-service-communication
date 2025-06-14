import { Flex, Heading, IconButton } from '@chakra-ui/react'
import { FaMoon, FaSun } from 'react-icons/fa'
import { useColorMode } from './ui/color-mode'

export const Header = () => {
    const { colorMode, toggleColorMode } = useColorMode()

    return (
        <Flex align="center" justify="space-between">
            <Heading size="2xl" color="teal.500">
                App Market
            </Heading>
            <IconButton
                aria-label="Toggle color mode"
                onClick={toggleColorMode}
                variant="ghost"
                size="lg"
                left={4}
            >
                {colorMode === 'light' ? <FaMoon /> : <FaSun />}
            </IconButton>
        </Flex>
    )
} 