import React from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';
import { typography } from '../theme/typography';

export const LoadingScreen: React.FC = () => {
  return (
    <View style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.appName}>UMG Bus</Text>
        <View style={styles.indicatorWrapper}>
          <ActivityIndicator size="small" color={colors.primary} />
        </View>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    alignItems: 'center',
  },
  appName: {
    ...typography.h1,
    color: colors.textPrimary,
    letterSpacing: -1,
  },
  indicatorWrapper: {
    marginTop: spacing.lg,
    opacity: 0.6,
  },
});
