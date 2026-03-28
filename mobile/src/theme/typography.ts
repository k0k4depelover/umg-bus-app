import { TextStyle, Platform } from 'react-native';

const fontFamily = Platform.select({
  ios: 'System',
  android: 'Roboto',
  default: 'System',
});

export const typography: Record<string, TextStyle> = {
  h1: { fontSize: 32, fontWeight: '700', lineHeight: 40, fontFamily, letterSpacing: -0.5 },
  h2: { fontSize: 24, fontWeight: '700', lineHeight: 32, fontFamily, letterSpacing: -0.3 },
  h3: { fontSize: 20, fontWeight: '600', lineHeight: 28, fontFamily },
  body: { fontSize: 16, fontWeight: '400', lineHeight: 24, fontFamily },
  bodyBold: { fontSize: 16, fontWeight: '600', lineHeight: 24, fontFamily },
  caption: { fontSize: 14, fontWeight: '400', lineHeight: 20, fontFamily },
  captionBold: { fontSize: 14, fontWeight: '600', lineHeight: 20, fontFamily },
  small: { fontSize: 12, fontWeight: '400', lineHeight: 16, fontFamily },
  button: { fontSize: 16, fontWeight: '600', lineHeight: 24, fontFamily, letterSpacing: 0.3 },
};
