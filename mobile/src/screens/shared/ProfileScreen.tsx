import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';
import { useAuth } from '../../hooks/useAuth';

export function ProfileScreen() {
  const { claims, logout } = useAuth();

  const handleLogout = () => {
    Alert.alert('Cerrar sesión', '¿Estás seguro que deseas cerrar sesión?', [
      { text: 'Cancelar', style: 'cancel' },
      { text: 'Cerrar sesión', style: 'destructive', onPress: logout },
    ]);
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        {/* Header */}
        <Text style={styles.title}>Perfil</Text>

        {/* Avatar placeholder */}
        <View style={styles.avatarContainer}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>
              {claims?.role === 'pilot' ? '🚌' : '🎓'}
            </Text>
          </View>
          <Text style={styles.roleLabel}>
            {claims?.role === 'pilot' ? 'Piloto' : 'Estudiante'}
          </Text>
        </View>

        {/* Info cards */}
        <View style={styles.card}>
          <View style={styles.cardRow}>
            <Text style={styles.cardLabel}>ID de usuario</Text>
            <Text style={styles.cardValue} numberOfLines={1}>
              {claims?.user_id?.slice(0, 8) || '—'}...
            </Text>
          </View>
          <View style={styles.divider} />
          <View style={styles.cardRow}>
            <Text style={styles.cardLabel}>Campus ID</Text>
            <Text style={styles.cardValue} numberOfLines={1}>
              {claims?.campus_id?.slice(0, 8) || '—'}...
            </Text>
          </View>
          <View style={styles.divider} />
          <View style={styles.cardRow}>
            <Text style={styles.cardLabel}>Rol</Text>
            <Text style={styles.cardValue}>
              {claims?.role === 'pilot' ? 'Piloto' : 'Estudiante'}
            </Text>
          </View>
        </View>

        {/* Placeholder for future features */}
        <View style={styles.placeholder}>
          <Text style={styles.placeholderIcon}>🚧</Text>
          <Text style={styles.placeholderText}>
            Más opciones de perfil próximamente
          </Text>
          <Text style={styles.placeholderSubtext}>
            Edición de nombre, foto de perfil y preferencias
          </Text>
        </View>

        {/* Logout */}
        <TouchableOpacity
          style={styles.logoutButton}
          onPress={handleLogout}
          activeOpacity={0.7}
        >
          <Text style={styles.logoutText}>Cerrar sesión</Text>
        </TouchableOpacity>
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
  avatarContainer: {
    alignItems: 'center',
    marginBottom: spacing.lg,
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: colors.primaryLight,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.sm,
  },
  avatarText: {
    fontSize: 36,
  },
  roleLabel: {
    ...typography.captionBold,
    color: colors.primary,
    backgroundColor: colors.primaryLight,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.xs,
    borderRadius: 20,
    overflow: 'hidden',
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
    marginBottom: spacing.lg,
  },
  cardRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.sm,
  },
  cardLabel: {
    ...typography.caption,
    color: colors.textSecondary,
  },
  cardValue: {
    ...typography.captionBold,
    color: colors.textPrimary,
    maxWidth: '50%',
  },
  divider: {
    height: 1,
    backgroundColor: colors.border,
  },
  placeholder: {
    alignItems: 'center',
    paddingVertical: spacing.xl,
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
  logoutButton: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.primary,
    borderRadius: 12,
    paddingVertical: 14,
    alignItems: 'center',
    marginTop: 'auto',
    marginBottom: spacing.lg,
  },
  logoutText: {
    ...typography.button,
    color: colors.primary,
  },
});
