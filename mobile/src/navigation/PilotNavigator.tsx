import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { PilotTabParamList } from '../types';
import { PilotMapScreen } from '../screens/pilot/PilotMapScreen';
import { ProfileScreen } from '../screens/shared/ProfileScreen';
import { SettingsScreen } from '../screens/shared/SettingsScreen';
import { colors } from '../theme/colors';
import { Text, StyleSheet } from 'react-native';

const Tab = createBottomTabNavigator<PilotTabParamList>();

function TabIcon({ label, focused }: { label: string; focused: boolean }) {
  const icons: Record<string, string> = {
    Mapa: '🗺️',
    Perfil: '👤',
    Ajustes: '⚙️',
  };
  return (
    <Text style={[styles.icon, focused && styles.iconFocused]}>
      {icons[label] || '•'}
    </Text>
  );
}

export function PilotNavigator() {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textTertiary,
        tabBarStyle: styles.tabBar,
        tabBarLabelStyle: styles.tabLabel,
        tabBarIcon: ({ focused }) => {
          const labels: Record<string, string> = {
            PilotMap: 'Mapa',
            PilotProfile: 'Perfil',
            PilotSettings: 'Ajustes',
          };
          return <TabIcon label={labels[route.name] || ''} focused={focused} />;
        },
      })}
    >
      <Tab.Screen
        name="PilotMap"
        component={PilotMapScreen}
        options={{ tabBarLabel: 'Mapa' }}
      />
      <Tab.Screen
        name="PilotProfile"
        component={ProfileScreen}
        options={{ tabBarLabel: 'Perfil' }}
      />
      <Tab.Screen
        name="PilotSettings"
        component={SettingsScreen}
        options={{ tabBarLabel: 'Ajustes' }}
      />
    </Tab.Navigator>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    backgroundColor: colors.surface,
    borderTopColor: colors.border,
    borderTopWidth: 1,
    height: 60,
    paddingBottom: 8,
    paddingTop: 4,
  },
  tabLabel: {
    fontSize: 12,
    fontWeight: '500',
  },
  icon: {
    fontSize: 20,
  },
  iconFocused: {
    transform: [{ scale: 1.1 }],
  },
});
