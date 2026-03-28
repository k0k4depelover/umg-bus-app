import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { StudentTabParamList } from '../types';
import { StudentMapScreen } from '../screens/student/StudentMapScreen';
import { ProfileScreen } from '../screens/shared/ProfileScreen';
import { SettingsScreen } from '../screens/shared/SettingsScreen';
import { CampusChangeScreen } from '../screens/shared/CampusChangeScreen';
import { colors } from '../theme/colors';
import { Text, StyleSheet } from 'react-native';

const Tab = createBottomTabNavigator<StudentTabParamList>();

function TabIcon({ label, focused }: { label: string; focused: boolean }) {
  const icons: Record<string, string> = {
    Mapa: '🗺️',
    Perfil: '👤',
    Ajustes: '⚙️',
    Campus: '🏫',
  };
  return (
    <Text style={[styles.icon, focused && styles.iconFocused]}>
      {icons[label] || '•'}
    </Text>
  );
}

export function StudentNavigator() {
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
            StudentMap: 'Mapa',
            StudentProfile: 'Perfil',
            StudentSettings: 'Ajustes',
            StudentCampus: 'Campus',
          };
          return <TabIcon label={labels[route.name] || ''} focused={focused} />;
        },
      })}
    >
      <Tab.Screen
        name="StudentMap"
        component={StudentMapScreen}
        options={{ tabBarLabel: 'Mapa' }}
      />
      <Tab.Screen
        name="StudentCampus"
        component={CampusChangeScreen}
        options={{ tabBarLabel: 'Campus' }}
      />
      <Tab.Screen
        name="StudentProfile"
        component={ProfileScreen}
        options={{ tabBarLabel: 'Perfil' }}
      />
      <Tab.Screen
        name="StudentSettings"
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
