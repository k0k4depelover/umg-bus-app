import React, { useState, useEffect, useCallback } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Alert } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import MapLibreGL from '@maplibre/maplibre-react-native';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';
import { useAuth } from '../../hooks/useAuth';
import { useWebSocket } from '../../hooks/useWebSocket';
import { campusService } from '../../api/campus';
import { Campus, LiveLocation } from '../../types';
import { StatusBadge } from '../../components/StatusBadge';

// Initialize MapLibre
MapLibreGL.setAccessToken(null);

export function StudentMapScreen() {
  const { claims, logout } = useAuth();
  const [campus, setCampus] = useState<Campus | null>(null);
  const [busLocation, setBusLocation] = useState<LiveLocation | null>(null);
  const [routeCoords, setRouteCoords] = useState<[number, number][]>([]);

  // Handle incoming bus location from WebSocket
  const handleBusUpdate = useCallback((data: LiveLocation) => {
    setBusLocation(data);
    // Build route trail from accumulated positions
    setRouteCoords((prev) => {
      const next = [...prev, [data.lng, data.lat] as [number, number]];
      // Keep last 200 points to avoid memory issues
      return next.length > 200 ? next.slice(-200) : next;
    });
  }, []);

  // WebSocket connection for receiving live bus locations
  const { isConnected, connect, disconnect } = useWebSocket({
    role: 'student',
    onMessage: handleBusUpdate,
    onConnected: () => console.log('Student WS connected'),
    onDisconnected: () => console.log('Student WS disconnected'),
    onError: (err) => console.warn('Student WS error:', err),
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

  // Connect to WebSocket on mount
  useEffect(() => {
    connect();
    return () => disconnect();
  }, [connect, disconnect]);

  // Also fetch initial live location via GraphQL
  useEffect(() => {
    if (claims?.campus_id) {
      campusService
        .getLiveLocation(claims.campus_id)
        .then((loc) => {
          if (loc) {
            setBusLocation({
              pilot_id: loc.pilotId,
              campus_id: loc.campusID,
              lat: loc.lat,
              lng: loc.lng,
              bearing: loc.bearing,
              speed: loc.speedKmh,
              updated_at: loc.updatedAt,
            } as any);
          }
        })
        .catch(() => {
          // No active pilot - that's fine
        });
    }
  }, [claims?.campus_id]);

  const handleLogout = () => {
    Alert.alert('Cerrar sesión', '¿Estás seguro?', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: 'Cerrar sesión',
        style: 'destructive',
        onPress: () => {
          disconnect();
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
    : [-90.5519, 14.5833];

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

  const routeGeoJSON =
    routeCoords.length >= 2
      ? {
          type: 'Feature' as const,
          properties: {},
          geometry: {
            type: 'LineString' as const,
            coordinates: routeCoords,
          },
        }
      : null;

  const hasPilot = busLocation !== null;
  const lastUpdate = busLocation?.updated_at
    ? new Date(busLocation.updated_at).toLocaleTimeString()
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
          animationMode="flyTo"
          animationDuration={1000}
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

        {/* Route trail */}
        {routeGeoJSON && (
          <MapLibreGL.ShapeSource id="route-trail" shape={routeGeoJSON}>
            <MapLibreGL.LineLayer
              id="route-trail-line"
              style={{
                lineColor: colors.routePath,
                lineWidth: 4,
                lineCap: 'round',
                lineJoin: 'round',
              }}
            />
          </MapLibreGL.ShapeSource>
        )}

        {/* Bus marker */}
        {busLocation && (
          <MapLibreGL.PointAnnotation
            id="bus-location"
            coordinate={[busLocation.lng, busLocation.lat]}
          >
            <View style={styles.busMarker}>
              <View style={styles.busMarkerInner}>
                <Text style={styles.busMarkerIcon}>🚌</Text>
              </View>
              <View style={styles.busMarkerPulse} />
            </View>
          </MapLibreGL.PointAnnotation>
        )}
      </MapLibreGL.MapView>

      {/* Top info bar */}
      <SafeAreaView style={styles.topBar} edges={['top']}>
        <View style={styles.topBarContent}>
          <View style={styles.topBarLeft}>
            <Text style={styles.campusName}>
              {campus?.name || 'Cargando...'}
            </Text>
            <Text style={styles.campusCity}>{campus?.city || ''}</Text>
          </View>
          <View style={styles.topBarRight}>
            <StatusBadge
              label={isConnected ? 'En vivo' : 'Desconectado'}
              variant={isConnected ? 'online' : 'offline'}
            />
            <TouchableOpacity onPress={handleLogout} style={styles.logoutBtn}>
              <Text style={styles.logoutText}>Salir</Text>
            </TouchableOpacity>
          </View>
        </View>
      </SafeAreaView>

      {/* Bottom info panel */}
      <SafeAreaView style={styles.bottomBar} edges={['bottom']}>
        <View style={styles.bottomContent}>
          {hasPilot ? (
            <>
              <View style={styles.pilotInfoRow}>
                <View style={styles.pilotInfo}>
                  <View style={styles.pilotDot} />
                  <Text style={styles.pilotLabel}>Bus activo</Text>
                </View>
                {lastUpdate && (
                  <Text style={styles.updateTime}>{lastUpdate}</Text>
                )}
              </View>
              <View style={styles.statsRow}>
                <View style={styles.stat}>
                  <Text style={styles.statValue}>
                    {busLocation.speed?.toFixed(1) || '0'}
                  </Text>
                  <Text style={styles.statLabel}>km/h</Text>
                </View>
                <View style={styles.stat}>
                  <Text style={styles.statValue}>
                    {busLocation.bearing?.toFixed(0) || '0'}°
                  </Text>
                  <Text style={styles.statLabel}>Dirección</Text>
                </View>
                <View style={styles.stat}>
                  <Text style={styles.statValue}>
                    {routeCoords.length}
                  </Text>
                  <Text style={styles.statLabel}>Puntos</Text>
                </View>
              </View>
            </>
          ) : (
            <View style={styles.noPilot}>
              <Text style={styles.noPilotIcon}>🚌</Text>
              <Text style={styles.noPilotText}>
                No hay buses activos en este momento
              </Text>
              <Text style={styles.noPilotSubtext}>
                Cuando un piloto inicie su ruta, verás su posición aquí
              </Text>
            </View>
          )}
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
  topBarLeft: {
    flex: 1,
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
  // Bus marker
  busMarker: {
    width: 52,
    height: 52,
    alignItems: 'center',
    justifyContent: 'center',
  },
  busMarkerInner: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 3,
    borderColor: colors.surface,
    shadowColor: colors.primary,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.4,
    shadowRadius: 6,
    elevation: 6,
    zIndex: 2,
  },
  busMarkerIcon: {
    fontSize: 22,
  },
  busMarkerPulse: {
    position: 'absolute',
    width: 52,
    height: 52,
    borderRadius: 26,
    backgroundColor: colors.primary,
    opacity: 0.2,
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
  pilotInfoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  pilotInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
  },
  pilotDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: colors.success,
  },
  pilotLabel: {
    ...typography.captionBold,
    color: colors.textPrimary,
  },
  updateTime: {
    ...typography.small,
    color: colors.textTertiary,
  },
  statsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingTop: spacing.sm,
    borderTopWidth: 1,
    borderTopColor: colors.border,
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
  // No pilot state
  noPilot: {
    alignItems: 'center',
    paddingVertical: spacing.md,
  },
  noPilotIcon: {
    fontSize: 36,
    marginBottom: spacing.sm,
  },
  noPilotText: {
    ...typography.bodyBold,
    color: colors.textPrimary,
    textAlign: 'center',
    marginBottom: spacing.xs,
  },
  noPilotSubtext: {
    ...typography.caption,
    color: colors.textTertiary,
    textAlign: 'center',
  },
});
