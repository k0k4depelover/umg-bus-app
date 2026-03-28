import React from 'react';
import { View, Text, StyleSheet, Switch } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';

export function SettingsScreen() {
  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>Ajustes</Text>

        {/* Mock settings */}
        <View style={styles.card}>
          <View style={styles.settingRow}>
            <View>
              <Text style={styles.settingLabel}>Notificaciones</Text>
              <Text style={styles.settingDescription}>
                Recibir alertas cuando un bus esté cerca
              </Text>
            </View>
            <Switch
              value={false}
              disabled
              trackColor={{ false: colors.border, true: colors.primaryMuted }}
              thumbColor={colors.textTertiary}
            />
          </View>

          <View style={styles.divider} />

          <View style={styles.settingRow}>
            <View>
              <Text style={styles.settingLabel}>Modo oscuro</Text>
              <Text style={styles.settingDescription}>
                Cambiar el tema de la aplicación
              </Text>
            </View>
            <Switch
              value={false}
              disabled
              trackColor={{ false: colors.border, true: colors.primaryMuted }}
              thumbColor={colors.textTertiary}
            />
          </View>

          <View style={styles.divider} />

          <View style={styles.settingRow}>
            <View>
              <Text style={styles.settingLabel}>Sonido</Text>
              <Text style={styles.settingDescription}>
                Sonido al recibir actualizaciones
              </Text>
            </View>
            <Switch
              value={false}
              disabled
              trackColor={{ false: colors.border, true: colors.primaryMuted }}
              thumbColor={colors.textTertiary}
            />
          </View>
        </View>

        {/* Coming soon */}
        <View style={styles.placeholder}>
          <Text style={styles.placeholderIcon}>🚧</Text>
          <Text style={styles.placeholderText}>
            Ajustes en desarrollo
          </Text>
          <Text style={styles.placeholderSubtext}>
            Estas opciones estarán disponibles en futuras versiones
          </Text>
        </View>

        {/* Version */}
        <Text style={styles.version}>UMG Bus v0.1.0 — Prototipo</Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  content: {
    flex: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
  },
  title: {
    ...typography.h2,
    color: colors.textPrimary,
    marginBottom: spacing.lg,
  },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 16,
    padding: spacing.md,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  settingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.sm,
  },
  settingLabel: {
    ...typography.body,
    color: colors.textPrimary,
  },
  settingDescription: {
    ...typography.small,
    color: colors.textTertiary,
    marginTop: 2,
  },
  divider: {
    height: 1,
    backgroundColor: colors.border,
    marginVertical: spacing.xs,
  },
  placeholder: {
    alignItems: 'center',
    paddingVertical: spacing.xxl,
  },
  placeholderIcon: {
    fontSize: 32,
    marginBottom: spacing.sm,
  },
  placeholderText: {
    ...typography.body,
    color: colors.textSecondary,
    textAlign: 'center',
    marginBottom: spacing.xs,
  },
  placeholderSubtext: {
    ...typography.caption,
    color: colors.textTertiary,
    textAlign: 'center',
  },
  version: {
    ...typography.small,
    color: colors.textTertiary,
    textAlign: 'center',
    marginTop: 'auto',
    marginBottom: spacing.lg,
  },
});
