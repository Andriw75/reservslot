import { type Type } from '@angular/core';
import { HomeIcon } from '../../../common/icons/home.icon';
import { TagIcon } from '../../../common/icons/tag.icon';
import { LayersIcon } from '../../../common/icons/layers.icon';
import { CalendarIcon } from '../../../common/icons/calendar.icon';

export interface IconConfig {
  component: Type<any>;
  inputs?: Record<string, any>;
}

export interface MenuItem {
  label: string;
  key: string;
  path?: string;
  icon?: IconConfig;
  requiredPermission?: string;
  submenu?: MenuItem[];
}

export const MENU: MenuItem[] = [
  {
    label: 'Inicio',
    key: 'home',
    path: '/dashboard',
    icon: { component: HomeIcon, inputs: { size: '20' } },
  },
  {
    label: 'Categorías',
    key: 'categorias',
    path: '/dashboard/categorias',
    icon: { component: TagIcon, inputs: { size: '20' } },
  },
  {
    label: 'Ambientes',
    key: 'ambientes',
    path: '/dashboard/ambientes',
    icon: { component: LayersIcon, inputs: { size: '20' } },
  },
  {
    label: 'Ocupaciones',
    key: 'ocupaciones',
    path: '/dashboard/ocupaciones',
    icon: { component: CalendarIcon, inputs: { size: '20' } },
  },
];
