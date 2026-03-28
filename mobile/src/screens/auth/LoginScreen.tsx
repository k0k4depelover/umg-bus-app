import React, { useState } from 'react';
import {
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Button } from '../../components/Button';
import { Input } from '../../components/Input';
import { useAuth } from '../../hooks/useAuth';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';

export function LoginScreen() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'student' | 'pilot'>('student');
  const [pilotCode, setPilotCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const { login, error, clearError } = useAuth();

  // Mejora 1: Limpieza integral al cambiar de rol
  const handleRoleChange = (newRole: 'student' | 'pilot') => {
    setRole(newRole);
    setPilotCode('');
    clearError();
  };

  const handleLogin = async () => {
    // Validaciones locales
    if (!username.trim() || !password.trim()) {
      Alert.alert('Campos requeridos', 'Ingresa tu usuario y contraseña.');
      return;
    }

    if (role === 'pilot' && !pilotCode.trim()) {
      Alert.alert('Código requerido', 'Ingresa el código de piloto para acceder.');
      return;
    }

    setIsLoading(true);
    clearError();

    try {
      // Mejora 2: Aseguramos el envío del pilotCode si es necesario
      await login(username.trim(), password.trim(), role);
    } catch (err) {
      // Mejora 3: Captura de errores inesperados (ej. timeout o crash de red)
      console.error("Login unexpected error:", err);
      Alert.alert('Error', 'Ocurrió un fallo inesperado al intentar iniciar sesión.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.flex}
      >
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          // Mejora 5: 'always' permite cerrar el teclado al tocar fuera de los inputs más fácilmente
          keyboardShouldPersistTaps="always"
          bounces={false}
        >
          {/* Header */}
          <View style={styles.header}>
            <View style={styles.logoContainer}>
              <View style={styles.logo}>
                <Text style={styles.logoText}>UMG</Text>
              </View>
            </View>
            <Text style={styles.title}>UMG Bus</Text>
            <Text style={styles.subtitle}>
              Monitoreo de transporte universitario
            </Text>
          </View>

          {/* Role Selector */}
          <View style={styles.roleSelector}>
            <TouchableOpacity
              style={[
                styles.roleButton,
                role === 'student' && styles.roleButtonActive,
              ]}
              onPress={() => handleRoleChange('student')}
              activeOpacity={0.7}
            >
              <Text
                style={[
                  styles.roleText,
                  role === 'student' && styles.roleTextActive,
                ]}
              >
                Estudiante
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[
                styles.roleButton,
                role === 'pilot' && styles.roleButtonActive,
              ]}
              onPress={() => handleRoleChange('pilot')}
              activeOpacity={0.7}
            >
              <Text
                style={[
                  styles.roleText,
                  role === 'pilot' && styles.roleTextActive,
                ]}
              >
                Piloto
              </Text>
            </TouchableOpacity>
          </View>

          {/* Form */}
          <View style={styles.form}>
          <Input
            label="Usuario"
            value={username}
            onChangeText={(val) => {
              setUsername(val);
              if (error) clearError();
            }}
            placeholder="Tu nombre de usuario"
            autoCapitalize="none"
          />

            <Input
              label="Contraseña"
              value={password}
              onChangeText={(val) => {
                setPassword(val);
                if (error) clearError();
              }}
              placeholder="Tu contraseña"
              secureTextEntry
            />

            {role === 'pilot' && (
              <Input
                label="Código de piloto"
                value={pilotCode}
                onChangeText={setPilotCode}
                placeholder="Código de acceso"
                autoCapitalize="characters"
              />
            )}

            {/* Mejora 6: Feedback visual inmediato del error del hook */}
            {error ? (
              <Text style={styles.errorText}>{error}</Text>
            ) : null}

            <Button
              title="Iniciar sesión"
              onPress={handleLogin}
              loading={isLoading}
              disabled={isLoading}
              variant="primary"
              style={styles.loginButton}
            />

            <TouchableOpacity style={styles.registerLink} disabled>
              <Text style={styles.registerText}>
                ¿No tienes cuenta?{' '}
                <Text style={styles.registerTextBold}>
                  Próximamente
                </Text>
              </Text>
            </TouchableOpacity>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  flex: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: 80,
    paddingBottom: spacing.xl,
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing.xl,
  },
  logoContainer: {
    marginBottom: spacing.md,
  },
  logo: {
    width: 72,
    height: 72,
    borderRadius: 20,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    shadowColor: colors.primary,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  logoText: {
    ...typography.h3,
    color: colors.textOnPrimary,
    fontWeight: '800',
    fontSize: 18,
    letterSpacing: 1,
  },
  title: {
    ...typography.h1,
    color: colors.textPrimary,
    marginBottom: spacing.xs,
  },
  subtitle: {
    ...typography.body,
    color: colors.textSecondary,
    textAlign: 'center',
  },
  roleSelector: {
    flexDirection: 'row',
    backgroundColor: colors.surfaceElevated,
    borderRadius: 12,
    padding: 4,
    marginBottom: spacing.lg,
  },
  roleButton: {
    flex: 1,
    paddingVertical: 12,
    alignItems: 'center',
    borderRadius: 10,
  },
  roleButtonActive: {
    backgroundColor: colors.surface,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 3,
    elevation: 2,
  },
  roleText: {
    ...typography.captionBold,
    color: colors.textTertiary,
  },
  roleTextActive: {
    color: colors.primary,
  },
  form: {
    gap: spacing.md,
  },
  errorText: {
    ...typography.caption,
    color: colors.error,
    textAlign: 'center',
    marginTop: -spacing.xs,
    paddingHorizontal: spacing.sm,
  },
  loginButton: {
    marginTop: spacing.sm,
  },
  registerLink: {
    alignItems: 'center',
    paddingVertical: spacing.md,
  },
  registerText: {
    ...typography.caption,
    color: colors.textTertiary,
  },
  registerTextBold: {
    color: colors.textTertiary,
    fontWeight: '600',
    textDecorationLine: 'underline',
  },
});
