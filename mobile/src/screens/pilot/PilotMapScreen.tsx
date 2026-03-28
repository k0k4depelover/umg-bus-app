import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  Animated,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import MapLibreGL from '@maplibre/maplibre-react-native';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';
import { useAuth } from '../../hooks/useAuth';
import { useLocation } from '../../hooks/useLocation';
import { useWebSocket } from '../../hooks/useWebSocket';
import { campusService } from '../../api/campus';
import { Campus, PilotPing, MapBounds } from '../../types';
import { StatusBadge } from '../../components/StatusBadge';

// Initialize MapLibre
MapLibreGL.setAccessToken(null);

export function PilotMapScreen() {
  const { claims, logout } = useAuth();
  const { location, isTracking, hasPermission, startTracking, stopTracking } =
    useLocation();
  const [campus, setCampus] = useState<Campus | null>(null);
  const [pingCount, setPingCount] = useState(0);
  const pulseAnim = useRef(new Animated.Value(1)).current;

  // WebSocket connection
  const { isConnected, connect, disconnect, sendPing } = useWebSocket({
    role: 'pilot',
    onConnected: () => console.log('Pilot WS connected'),
    onDisconnected: () => console.log('Pilot WS disconnected'),
    onError: (err) => console.warn('Pilot WS error:', err),
  });

  // Load campus data
  useEffect(() => {
    if (claims?.campus_id) {
      campusService
        .getCampus(claims.campus_id)
        .then(setCampus)
        .catch((err) => console.warn('Failed to load campus:', err));
    }
  }, [claims?.campus_id]);

  // Pulse animation when tracking
  useEffect(() => {
    if (isTracking) {
      const pulse = Animated.loop(
        Animated.sequence([
          Animated.timing(pulseAnim, {
            toValue: 1.2,
            duration: 1000,
            useNativeDriver: true,
          }),
          Animated.timing(pulseAnim, {
            toValue: 1,
            duration: 1000,
            useNativeDriver: true,
          }),
        ])
      );
      pulse.start();
      return () => pulse.stop();
    } else {
      pulseAnim.setValue(1);
    }
  }, [isTracking, pulseAnim]);

  // Handle location updates — send via WebSocket
  const handleLocationUpdate = useCallback(
    (ping: PilotPing) => {
      sendPing(ping);
      setPingCount((c) => c + 1);
    },
    [sendPing]
  );

  const handleToggleTracking = () => {
    if (isTracking) {
      stopTracking();
      disconnect();
    } else {
      if (!hasPermission) {
        Alert.alert(
          'Permiso requerido',
          'Necesitas otorgar permiso de ubicación para transmitir.'
        );
        return;
      }
      connect();
      startTracking(handleLocationUpdate);
    }
  };

  const handleLogout = () => {
    Alert.alert('Cerrar sesión', '¿Estás seguro?', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: 'Cerrar sesión',
        style: 'destructive',
        onPress: () => {
          if (isTracking) {
            stopTracking();
            disconnect();
          }
          logout();
        },
      },
    ]);
  };

  const boundsCenter = campus?.bounds
    ? [
        (campus.bounds.swLng + campus.bounds.neLng) / 2,
        (campus.bounds.swLat + campus.bounds.neLat) / 2,
      ]
    : [-90.5519, 14.5833]; // Guatemala default

  const campusBoundsGeoJSON = campus?.bounds
    ? {
        type: 'Feature' as const,
        properties: {},
        geometry: {
          type: 'Polygon' as const,
          coordinates: [
            [
              [campus.bounds.swLng, campus.bounds.swLat],
              [campus.bounds.neLng, campus.bounds.swLat],
              [campus.bounds.neLng, campus.bounds.neLat],
              [campus.bounds.swLng, campus.bounds.neLat],
              [campus.bounds.swLng, campus.bounds.swLat],
            ],
          ],
        },
      }
    : null;

  return (
    <View style={styles.container}>
      {/* Map */}
      <MapLibreGL.MapView
        style={styles.map}
        styleURL="https://demotiles.maplibre.org/style.json"
        logoEnabled={false}
        attributionEnabled={false}
      >
        <MapLibreGL.Camera
          defaultSettings={{
            centerCoordinate: boundsCenter,
            zoomLevel: 15,
          }}
        />

        {/* Campus bounds overlay */}
        {campusBoundsGeoJSON && (
          <MapLibreGL.ShapeSource
            id="campus-bounds"
            shape={campusBoundsGeoJSON}
          >
            <MapLibreGL.FillLayer
              id="campus-bounds-fill"
              style={{
                fillColor: colors.campusBounds,
                fillOutlineColor: colors.campusBorder,
              }}
            />
            <MapLibreGL.LineLayer
              id="campus-bounds-border"
              style={{
                lineColor: colors.campusBorder,
                lineWidth: 2,
                lineDasharray: [4, 2],
              }}
            />
          </MapLibreGL.ShapeSource>
        )}

        {/* Pilot's current position */}
        {location && (
          <MapLibreGL.PointAnnotation
            id="pilot-location"
            coordinate={[location.longitude, location.latitude]}
          >
            <View style={styles.pilotMarker}>
              <View style={styles.pilotMarkerInner}>
                <Text style={styles.pilotMarkerIcon}>🚌</Text>
              </View>
            </View>
          </MapLibreGL.PointAnnotation>
        )}
      </MapLibreGL.MapView>

      {/* Top info bar */}
      <SafeAreaView style={styles.topBar} edges={['top']}>
        <View style={styles.topBarContent}>
          <View>
            <Text style={styles.campusName}>
              {campus?.name || 'Cargando campus...'}
            </Text>
            <Text style={styles.campusCity}>{campus?.city || ''}</Text>
          </View>
          <View style={styles.topBarRight}>
            <StatusBadge
              label={isConnected ? 'Conectado' : 'Desconectado'}
              variant={isConnected ? 'online' : 'offline'}
            />
            <TouchableOpacity onPress={handleLogout} style={styles.logoutBtn}>
              <Text style={styles.logoutText}>Salir</Text>
            </TouchableOpacity>
          </View>
        </View>
      </SafeAreaView>

      {/* Bottom controls */}
      <SafeAreaView style={styles.bottomBar} edges={['bottom']}>
        <View style={styles.bottomContent}>
          {/* Stats */}
          <View style={styles.statsRow}>
            <View style={styles.stat}>
              <Text style={styles.statValue}>
                {location ? `${location.speed.toFixed(1)}` : '--'}
              </Text>
              <Text style={styles.statLabel}>km/h</Text>
            </View>
            <View style={styles.stat}>
              <Text style={styles.statValue}>
                {location ? `${location.bearing.toFixed(0)}°` : '--'}
              </Text>
              <Text style={styles.statLabel}>Dirección</Text>
            </View>
            <View style={styles.stat}>
              <Text style={styles.statValue}>{pingCount}</Text>
              <Text style={styles.statLabel}>Pings</Text>
            </View>
          </View>

          {/* Toggle button */}
          <TouchableOpacity
            style={[
              styles.trackingButton,
              isTracking && styles.trackingButtonActive,
            ]}
            onPress={handleToggleTracking}
            activeOpacity={0.8}
          >
            <Animated.View
              style={[
                styles.trackingButtonInner,
                isTracking && styles.trackingButtonInnerActive,
                { transform: [{ scale: pulseAnim }] },
              ]}
            >
              <Text style={styles.trackingButtonIcon}>
                {isTracking ? '⏹' : '▶️'}
              </Text>
            </Animated.View>
            <Text
              style={[
                styles.trackingButtonText,
                isTracking && styles.trackingButtonTextActive,
              ]}
            >
              {isTracking ? 'Detener transmisión' : 'Iniciar transmisión'}
            </Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  map: {
    flex: 1,
  },
  // Top bar
  topBar: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
  },
  topBarContent: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    marginHorizontal: spacing.md,
    marginTop: spacing.sm,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
  },
  campusName: {
    ...typography.bodyBold,
    color: colors.textPrimary,
  },
  campusCity: {
    ...typography.small,
    color: colors.textSecondary,
  },
  topBarRight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  logoutBtn: {
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs,
  },
  logoutText: {
    ...typography.caption,
    color: colors.primary,
    fontWeight: '600',
  },
  // Pilot marker
  pilotMarker: {
    width: 48,
    height: 48,
    alignItems: 'center',
    justifyContent: 'center',
  },
  pilotMarkerInner: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 3,
    borderColor: colors.surface,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.25,
    shadowRadius: 4,
    elevation: 4,
  },
  pilotMarkerIcon: {
    fontSize: 20,
  },
  // Bottom bar
  bottomBar: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
  },
  bottomContent: {
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    marginHorizontal: spacing.md,
    marginBottom: spacing.md,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.md,
    borderRadius: 20,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: -2 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
  },
  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginBottom: spacing.md,
  },
  stat: {
    alignItems: 'center',
  },
  statValue: {
    ...typography.h3,
    color: colors.textPrimary,
  },
  statLabel: {
    ...typography.small,
    color: colors.textTertiary,
  },
  // Tracking button
  trackingButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.primaryLight,
    borderRadius: 16,
    paddingVertical: 14,
    paddingHorizontal: spacing.lg,
    gap: spacing.sm,
  },
  trackingButtonActive: {
    backgroundColor: colors.primary,
  },
  trackingButtonInner: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  trackingButtonInnerActive: {
    backgroundColor: colors.surface,
  },
  trackingButtonIcon: {
    fontSize: 16,
  },
  trackingButtonText: {
    ...typography.button,
    color: colors.primary,
  },
  trackingButtonTextActive: {
    color: colors.textOnPrimary,
  },
});
