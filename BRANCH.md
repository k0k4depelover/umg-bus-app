# Branches en Git

## Que es una branch?

Una branch (rama) es una linea independiente de desarrollo. Permite trabajar en features, fixes o experimentos sin afectar el codigo principal.

## Ramas principales

| Rama | Proposito |
|------|-----------|
| `main` | Codigo estable y listo para produccion |
| `develop` | Integracion de features antes de pasar a main (opcional) |

## Ramas de trabajo

| Prefijo | Uso | Ejemplo |
|---------|-----|---------|
| `feature/` | Nueva funcionalidad | `feature/gps-tracking` |
| `fix/` | Correccion de bugs | `fix/websocket-connection` |
| `hotfix/` | Fix urgente en produccion | `hotfix/crash-on-login` |
| `refactor/` | Reestructurar sin cambiar funcionalidad | `refactor/hub-cleanup` |

## Comandos basicos

```bash
# Ver todas las ramas
git branch -a

# Crear y cambiar a nueva rama
git checkout -b feature/mi-feature

# Cambiar entre ramas
git checkout main
git checkout feature/mi-feature

# Subir rama al remoto
git push -u origin feature/mi-feature

# Traer cambios de main a tu rama
git pull origin main

# Eliminar rama local (despues de merge)
git branch -d feature/mi-feature
```

## Flujo de trabajo

```
main ────────────────────────────────────────►
       \                          /
        feature/gps-tracking ────► (merge via PR)
```

1. Crear rama desde `main`: `git checkout -b feature/nueva-feature`
2. Hacer cambios y commits en esa rama
3. Push al remoto: `git push -u origin feature/nueva-feature`
4. Crear Pull Request en GitHub
5. Review y merge a `main`
6. Eliminar la rama

## Por que usar branches?

- **Aislamiento**: cambios no afectan a otros hasta hacer merge
- **Colaboracion**: cada quien trabaja en su rama sin conflictos
- **Historial limpio**: cada feature/fix tiene su propio contexto
- **Rollback facil**: si algo sale mal, la rama se descarta sin afectar main
