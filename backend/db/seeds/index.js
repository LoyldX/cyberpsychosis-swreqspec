/**
 * Database Seed Script
 * Initializes test data for local development
 *
 * Usage: npm run db:seed
 */

const { PrismaClient } = require('@prisma/client');

const prisma = new PrismaClient();

async function seed() {
  console.log('🌱 Seeding database...');

  try {
    // Create test members
    const member1 = await prisma.member.create({
      data: {
        studentId: 'STU001',
        name: 'John Doe',
        email: 'john@example.com',
        phone: '0812345678',
      },
    });

    const member2 = await prisma.member.create({
      data: {
        studentId: 'STU002',
        name: 'Jane Smith',
        email: 'jane@example.com',
        phone: '0887654321',
      },
    });

    console.log(`✓ Created ${2} members`);

    // Create test trainers
    const trainer1 = await prisma.trainer.create({
      data: {
        name: 'Alex Coach',
        email: 'alex@gym.local',
        phone: '0899999999',
      },
    });

    const trainer2 = await prisma.trainer.create({
      data: {
        name: 'Sarah Fitness',
        email: 'sarah@gym.local',
        phone: '0888888888',
      },
    });

    console.log(`✓ Created ${2} trainers`);

    // Create trainer shifts (recurring schedule)
    // Monday-Friday 09:00-10:00
    const daysOfWeek = [1, 2, 3, 4, 5]; // Monday to Friday
    for (const day of daysOfWeek) {
      await prisma.trainerShift.create({
        data: {
          trainerId: trainer1.id,
          dayOfWeek: day,
          startTime: '09:00',
          endTime: '10:00',
        },
      });

      await prisma.trainerShift.create({
        data: {
          trainerId: trainer2.id,
          dayOfWeek: day,
          startTime: '17:00',
          endTime: '18:00',
        },
      });
    }

    console.log(`✓ Created trainer shifts (10 total)`);

    // Create test classes (scheduled instances)
    // Class this coming Monday at 09:00
    const nextMonday = new Date();
    nextMonday.setDate(nextMonday.getDate() + ((1 - nextMonday.getDay() + 7) % 7));
    nextMonday.setHours(9, 0, 0, 0);

    const class1 = await prisma.class.create({
      data: {
        trainerId: trainer1.id,
        scheduledAt: nextMonday,
        durationMins: 60,
        capacity: 1,
        createdBy: 1,
      },
    });

    // Another class Wednesday at 09:00
    const wednesday = new Date(nextMonday);
    wednesday.setDate(wednesday.getDate() + 2);

    const class2 = await prisma.class.create({
      data: {
        trainerId: trainer1.id,
        scheduledAt: wednesday,
        durationMins: 60,
        capacity: 1,
        createdBy: 1,
      },
    });

    // Class from trainer2 Monday at 17:00
    const class3 = await prisma.class.create({
      data: {
        trainerId: trainer2.id,
        scheduledAt: nextMonday.getTime() + 8 * 3600 * 1000, // Monday 17:00
        durationMins: 60,
        capacity: 1,
      },
    });

    console.log(`✓ Created ${3} classes`);

    // Create test bookings
    const booking1 = await prisma.booking.create({
      data: {
        memberId: member1.id,
        classId: class1.id,
        status: 'confirmed',
      },
    });

    console.log(`✓ Created ${1} booking`);

    // Create inventory items
    const item1 = await prisma.inventoryItem.create({
      data: {
        code: 'DUMB-10KG',
        name: '10kg Dumbbell',
        category: 'equipment',
        quantity: 5,
        reorderPoint: 2,
        reorderQuantity: 5,
        unitPrice: 500,
      },
    });

    const item2 = await prisma.inventoryItem.create({
      data: {
        code: 'TOWEL-001',
        name: 'Gym Towel',
        category: 'consumable',
        quantity: 50,
        reorderPoint: 10,
        reorderQuantity: 20,
        unitPrice: 50,
      },
    });

    console.log(`✓ Created ${2} inventory items`);

    // Create occupancy snapshot
    await prisma.occupancySnapshot.create({
      data: {
        currentCount: 0,
        capacity: 100,
        utilizationRate: 0,
        lastUpdated: new Date(),
      },
    });

    console.log(`✓ Created occupancy snapshot`);

    console.log('\n✅ Database seeding complete!\n');
    console.log('Test data:');
    console.log(`  Members: ${member1.studentId}, ${member2.studentId}`);
    console.log(`  Trainers: ${trainer1.name}, ${trainer2.name}`);
    console.log(`  Classes: ${3} scheduled`);
    console.log(`  Bookings: ${1} confirmed`);
    console.log('\nYou can now test login with:');
    console.log(`  student_id: STU001`);
    console.log(`  password: demo123 (temporary - change in spec)`);
  } catch (error) {
    console.error('❌ Seeding failed:', error);
    process.exit(1);
  } finally {
    await prisma.$disconnect();
  }
}

seed();
