import React, { useState, useRef, useEffect } from 'react';
import {
  View,
  TextInput,
  Text,
  StyleSheet,
  Animated,
  ViewStyle,
  TextInputProps,
} from 'react-native';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';
import { typography } from '../theme/typography';

interface InputProps {
  label: string;
  value: string;
  onChangeText: (text: string) => void;
  placeholder?: string;
  secureTextEntry?: boolean;
  error?: string;
  icon?: React.ReactNode;
  autoCapitalize?: TextInputProps['autoCapitalize'];
  keyboardType?: TextInputProps['keyboardType'];
}

export const Input: React.FC<InputProps> = ({
  label,
  value,
  onChangeText,
  placeholder,
  secureTextEntry = false,
  error,
  icon,
  autoCapitalize,
  keyboardType,
}) => {
  const [isFocused, setIsFocused] = useState(false);
  const labelAnim = useRef(new Animated.Value(value ? 1 : 0)).current;

  const isActive = isFocused || value.length > 0;

  useEffect(() => {
    Animated.timing(labelAnim, {
      toValue: isActive ? 1 : 0,
      duration: 150,
      useNativeDriver: false,
    }).start();
  }, [isActive, labelAnim]);

  const labelTop = labelAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [14, 4],
  });

  const labelSize = labelAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [16, 12],
  });

  const borderColor = error
    ? colors.error
    : isFocused
      ? colors.borderFocused
      : colors.border;

  return (
    <View style={styles.wrapper}>
      <View
        style={[
          styles.container,
          { borderColor },
          isFocused && !error && styles.containerFocused,
          error ? styles.containerError : undefined,
        ]}
      >
        {icon && <View style={styles.iconContainer}>{icon}</View>}

        <View style={styles.inputWrapper}>
          <Animated.Text
            style={[
              styles.label,
              {
                top: labelTop,
                fontSize: labelSize,
                color: error
                  ? colors.error
                  : isFocused
                    ? colors.primary
                    : colors.textTertiary,
              },
            ]}
            numberOfLines={1}
          >
            {label}
          </Animated.Text>

          <TextInput
            style={[styles.input, isActive && styles.inputActive]}
            value={value}
            onChangeText={onChangeText}
            placeholder={isActive ? placeholder : undefined}
            placeholderTextColor={colors.textTertiary}
            secureTextEntry={secureTextEntry}
            autoCapitalize={autoCapitalize}
            keyboardType={keyboardType}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            selectionColor={colors.primary}
          />
        </View>
      </View>

      {error ? <Text style={styles.errorText}>{error}</Text> : null}
    </View>
  );
};

const styles = StyleSheet.create({
  wrapper: {
    marginBottom: spacing.sm,
  },
  container: {
    height: 52,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderWidth: 1.5,
    borderRadius: 12,
    paddingHorizontal: spacing.md,
  },
  containerFocused: {
    shadowColor: colors.primary,
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.08,
    shadowRadius: 8,
    elevation: 2,
  },
  containerError: {
    backgroundColor: '#FEF2F2',
  },
  iconContainer: {
    marginRight: spacing.sm,
    justifyContent: 'center',
    alignItems: 'center',
  },
  inputWrapper: {
    flex: 1,
    justifyContent: 'center',
    height: '100%',
  },
  label: {
    position: 'absolute',
    left: 0,
    fontWeight: '500',
  },
  input: {
    ...typography.body,
    color: colors.textPrimary,
    paddingTop: 12,
    paddingBottom: 0,
    height: '100%',
  },
  inputActive: {
    paddingTop: 16,
  },
  errorText: {
    ...typography.small,
    color: colors.error,
    marginTop: spacing.xs,
    marginLeft: spacing.xs,
  },
});
