import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { useAuth } from '../hooks/useAuth';
import { RootStackParamList } from '../types';
import { LoadingScreen } from '../components';
import { AuthNavigator } from './AuthNavigator';
import { PilotNavigator } from './PilotNavigator';
import { StudentNavigator } from './StudentNavigator';

const Stack = createNativeStackNavigator<RootStackParamList>();

export function RootNavigator() {
  const { isLoading, isAuthenticated, claims } = useAuth();

  if (isLoading) {
    return <LoadingScreen />;
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false, animation: 'fade' }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Auth" component={AuthNavigator} />
        ) : claims?.role === 'pilot' ? (
          <Stack.Screen name="PilotMain" component={PilotNavigator} />
        ) : (
          <Stack.Screen name="StudentMain" component={StudentNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
