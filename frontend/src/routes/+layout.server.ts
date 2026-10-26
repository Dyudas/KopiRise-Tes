import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
  return {
    cafe: {
      name: 'KopiRise',
      tagline: 'Cafe & Resto Jember',
      address: 'Jl. Semeru, Kloncing, Karangrejo, Kec. Sumbersari, Kabupaten Jember, Jawa Timur 68124',
      phone: '+62 812-3456-7890',
      email: 'kontak@kopirise.id',
      operatingHours: '08:00 - 22:00 WIB',
      currency: 'IDR'
    }
  };
};
