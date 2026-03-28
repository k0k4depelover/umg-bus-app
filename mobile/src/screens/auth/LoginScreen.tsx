import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';
import { Button } from '../../components/Button';
import { Input } from '../../components/Input';
import { useAuth } from '../../hooks/useAuth';

export function LoginScreen() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'student' | 'pilot'>('student');
  const [pilotCode, setPilotCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const { login, error, clearError } = useAuth();

  const handleLogin = async () => {
    if (!username.trim() || !password.trim()) {
      Alert.alert('Campos requeridos', 'Ingresa tu usuario y contraseña.');
      return;
    }

    if (role === 'pilot' && !pilotCode.trim()) {
      Alert.alert('Código requerido', 'Ingresa el código de piloto para acceder como conductor.');
      return;
    }

    setIsLoading(true);
    clearError();

    try {
      await login(username.trim(), password.trim(), role);
    } catch {
      // Error is handled by the store
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView style={styles.container}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.flex}
      >
        <ScrollView
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
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
              onPress={() => setRole('student')}
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
              onPress={() => setRole('pilot')}
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
              onChangeText={setUsername}
              placeholder="Tu nombre de usuario"
              autoCapitalize="none"
            />

            <Input
              label="Contraseña"
              value={password}
              onChangeText={setPassword}
              placeholder="Tu contraseña"
              secureTextEntry
            />

            {role === 'pilot' && (
              <Input
                label="Código de piloto"
                value={pilotCode}
                onChangeText={setPilotCode}
                placeholder="Código de acceso"
                autoCapitalize="none"
              />
            )}

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

            {/* Future registration link */}
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
    justifyContent: 'center',
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
