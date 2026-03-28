import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';
import { useAuth } from '../../hooks/useAuth';
import { campusService } from '../../api/campus';
import { Campus } from '../../types';

export function CampusChangeScreen() {
  const { claims } = useAuth();
  const [campuses, setCampuses] = useState<Campus[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    campusService
      .getCampuses()
      .then(setCampuses)
      .catch((err) => {
        console.warn('Failed to load campuses:', err);
        Alert.alert('Error', 'No se pudieron cargar los campus.');
      })
      .finally(() => setLoading(false));
  }, []);

  const handleChangeCampus = (campus: Campus) => {
    if (campus.campusID === claims?.campus_id) {
      Alert.alert('Campus actual', 'Ya estás en este campus.');
      return;
    }
    Alert.alert(
      'Cambiar campus',
      `¿Deseas cambiar al campus "${campus.name}" en ${campus.city}?`,
      [
        { text: 'Cancelar', style: 'cancel' },
        {
          text: 'Cambiar',
          onPress: () => {
            // TODO: Implement changeCampus mutation
            Alert.alert(
              'Próximamente',
              'El cambio de campus estará disponible pronto.'
            );
          },
        },
      ]
    );
  };

  const renderCampus = ({ item }: { item: Campus }) => {
    const isCurrent = item.campusID === claims?.campus_id;
    return (
      <TouchableOpacity
        style={[styles.campusCard, isCurrent && styles.campusCardCurrent]}
        onPress={() => handleChangeCampus(item)}
        activeOpacity={0.7}
      >
        <View style={styles.campusInfo}>
          <Text style={styles.campusName}>{item.name}</Text>
          <Text style={styles.campusCity}>{item.city}</Text>
          {item.activePilot && (
            <View style={styles.pilotRow}>
              <View
                style={[
                  styles.pilotDot,
                  item.activePilot.isOnline && styles.pilotDotOnline,
                ]}
              />
              <Text style={styles.pilotName}>
                {item.activePilot.fullName}
              </Text>
            </View>
          )}
        </View>
        {isCurrent && (
          <View style={styles.currentBadge}>
            <Text style={styles.currentBadgeText}>Actual</Text>
          </View>
        )}
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <Text style={styles.title}>Campus</Text>
        <Text style={styles.subtitle}>
          Selecciona el campus que deseas monitorear
        </Text>

        {loading ? (
          <View style={styles.loadingContainer}>
            <ActivityIndicator size="large" color={colors.primary} />
          </View>
        ) : (
          <FlatList
            data={campuses}
            keyExtractor={(item) => item.campusID}
            renderItem={renderCampus}
            contentContainerStyle={styles.list}
            ListEmptyComponent={
              <View style={styles.empty}>
                <Text style={styles.emptyIcon}>🏫</Text>
                <Text style={styles.emptyText}>
                  No hay campus disponibles
                </Text>
              </View>
            }
          />
        )}
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
    paddingTop: spacing.lg,
  },
  title: {
    ...typography.h2,
    color: colors.textPrimary,
    paddingHorizontal: spacing.lg,
  },
  subtitle: {
    ...typography.caption,
    color: colors.textSecondary,
    paddingHorizontal: spacing.lg,
    marginBottom: spacing.md,
  },
  list: {
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
  },
  campusCard: {
    backgroundColor: colors.surface,
    borderRadius: 16,
    padding: spacing.md,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
    marginBottom: spacing.sm,
  },
  campusCardCurrent: {
    borderWidth: 2,
    borderColor: colors.primary,
  },
  campusInfo: {
    flex: 1,
  },
  campusName: {
    ...typography.bodyBold,
    color: colors.textPrimary,
  },
  campusCity: {
    ...typography.caption,
    color: colors.textSecondary,
    marginTop: 2,
  },
  pilotRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    marginTop: spacing.xs,
  },
  pilotDot: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: colors.textTertiary,
  },
  pilotDotOnline: {
    backgroundColor: colors.success,
  },
  pilotName: {
    ...typography.small,
    color: colors.textSecondary,
  },
  currentBadge: {
    backgroundColor: colors.primaryLight,
    borderRadius: 12,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
  },
  currentBadgeText: {
    ...typography.small,
    color: colors.primary,
    fontWeight: '600',
  },
  loadingContainer: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
  },
  empty: {
    alignItems: 'center',
    paddingVertical: spacing.xxl,
  },
  emptyIcon: {
    fontSize: 48,
    marginBottom: spacing.md,
  },
  emptyText: {
    ...typography.body,
    color: colors.textSecondary,
  },
});
