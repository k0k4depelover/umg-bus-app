import React from 'react';
import { View, Text, StyleSheet, Platform } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';
import { typography } from '../theme/typography';
import { StatusBadge } from './StatusBadge';

interface MapHeaderProps {
  campusName: string;
  pilotName?: string;
  isOnline?: boolean;
  speed?: number;
}

export const MapHeader: React.FC<MapHeaderProps> = ({
  campusName,
  pilotName,
  isOnline = false,
  speed,
}) => {
  const insets = useSafeAreaInsets();

  return (
    <View style={[styles.container, { paddingTop: insets.top + spacing.sm }]}>
      <View style={styles.row}>
        <View style={styles.info}>
          <Text style={styles.campusName}>{campusName}</Text>
          {pilotName && (
            <Text style={styles.pilotName}>{pilotName}</Text>
          )}
        </View>

        <View style={styles.statusArea}>
          <StatusBadge
            label={isOnline ? 'Online' : 'Offline'}
            variant={isOnline ? 'online' : 'offline'}
          />
          {speed !== undefined && isOnline && (
            <Text style={styles.speed}>{speed.toFixed(0)} km/h</Text>
          )}
        </View>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    zIndex: 10,
    backgroundColor: 'rgba(255, 255, 255, 0.92)',
    paddingHorizontal: spacing.md,
    paddingBottom: spacing.md,
    borderBottomLeftRadius: 20,
    borderBottomRightRadius: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 8,
    elevation: 4,
    ...Platform.select({
      ios: {
        // On iOS the translucent background already provides a frosted look
      },
      android: {
        backgroundColor: 'rgba(255, 255, 255, 0.97)',
      },
    }),
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  info: {
    flex: 1,
    marginRight: spacing.sm,
  },
  campusName: {
    ...typography.h3,
    color: colors.textPrimary,
  },
  pilotName: {
    ...typography.caption,
    color: colors.textSecondary,
    marginTop: 2,
  },
  statusArea: {
    alignItems: 'flex-end',
    gap: spacing.xs,
  },
  speed: {
    ...typography.small,
    color: colors.textSecondary,
    fontWeight: '600',
  },
});
