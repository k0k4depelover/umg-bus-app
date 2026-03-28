import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';
import { typography } from '../theme/typography';

type BadgeVariant = 'online' | 'offline' | 'tracking' | 'idle';

interface StatusBadgeProps {
  label: string;
  variant: BadgeVariant;
}

const variantStyles: Record<
  BadgeVariant,
  { bg: string; text: string; dot: string }
> = {
  online: {
    bg: '#F0FDF4',
    text: '#15803D',
    dot: '#16A34A',
  },
  tracking: {
    bg: '#F0FDF4',
    text: '#15803D',
    dot: '#16A34A',
  },
  offline: {
    bg: colors.surfaceElevated,
    text: colors.textTertiary,
    dot: colors.textTertiary,
  },
  idle: {
    bg: colors.surfaceElevated,
    text: colors.textTertiary,
    dot: colors.textTertiary,
  },
};

export const StatusBadge: React.FC<StatusBadgeProps> = ({ label, variant }) => {
  const theme = variantStyles[variant];

  return (
    <View style={[styles.container, { backgroundColor: theme.bg }]}>
      <View style={[styles.dot, { backgroundColor: theme.dot }]} />
      <Text style={[styles.text, { color: theme.text }]}>{label}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    alignSelf: 'flex-start',
    paddingHorizontal: spacing.sm + 4,
    paddingVertical: spacing.xs + 2,
    borderRadius: 999,
    gap: spacing.xs + 2,
  },
  dot: {
    width: 7,
    height: 7,
    borderRadius: 4,
  },
  text: {
    ...typography.small,
    fontWeight: '600',
  },
});
