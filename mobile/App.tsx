import React, { useEffect } from 'react';
import { StatusBar, LogBox } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { RootNavigator } from './src/navigation';
import { useAuth } from './src/hooks/useAuth';
import { colors } from './src/theme/colors';

// Suppress known warnings in dev
LogBox.ignoreLogs([
  'Non-serializable values were found in the navigation state',
]);

function AppContent() {
  const init = useAuth((state) => state.init);

  useEffect(() => {
    init();
  }, [init]);

  return (
    <>
      <StatusBar
        barStyle="dark-content"
        backgroundColor={colors.background}
        translucent={false}
      />
      <RootNavigator />
    </>
  );
}

function App() {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <SafeAreaProvider>
        <AppContent />
      </SafeAreaProvider>
    </GestureHandlerRootView>
  );
}

export default App;
