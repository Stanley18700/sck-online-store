import { Test, TestingModule } from '@nestjs/testing';
import { getRepositoryToken } from '@nestjs/typeorm';
import { CreatePointDto } from '../point.dto';
import { Point } from '../point.entity';
import { PointService } from '../point.service';

describe('PointService', () => {
  let service: PointService;

  const mockPointRepository = {
    save: jest.fn(),
    find: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        PointService,
        {
          provide: getRepositoryToken(Point),
          useValue: mockPointRepository,
        },
      ],
    }).compile();

    service = module.get<PointService>(PointService);
  });

  it('Should be defined', () => {
    expect(service).toBeDefined();
  });

  it('Create => Should create a new point and return its data', async () => {
    // arrange
    const createPointInput = {
      orgId: 1,
      userId: 1,
      amount: 200,
    } as CreatePointDto;

    const createPointResponse = {
      id: 1,
      orgId: 1,
      userId: 1,
      amount: 200,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    } as CreatePointDto;

    jest
      .spyOn(mockPointRepository, 'save')
      .mockReturnValue(createPointResponse);

    // act
    const result = await service.deductPoint(createPointInput);

    // assert
    expect(mockPointRepository.save).toBeCalled();
    expect(mockPointRepository.save).toBeCalledWith(createPointInput);
    expect(result).toEqual(createPointResponse);
  });

  it('Calculate => Should return 20 points from amount 1000', () => {
    // arrange
    const amount = 1000;
    const expected = 20;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 21 points from amount 1060', () => {
    // arrange
    const amount = 1060;
    const expected = 21;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 1 point from amount 60', () => {
    // arrange
    const amount = 60;
    const expected = 1;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 0 points from amount 0', () => {
    // arrange
    const amount = 0;
    const expected = 0;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 2 points from amount 100', () => {
    // arrange
    const amount = 100;
    const expected = 2;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 11 points from amount 599.999', () => {
    // arrange
    const amount = 599.999;
    const expected = 11;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Calculate => Should return 0 points from amount -599.999', () => {
    // arrange
    const amount = -599.999;
    const expected = 0;

    // act
    const result = service.calculatePoint(amount);

    // assert
    expect(result).toEqual(expected);
  });

  it('Find => should return an array of point', async () => {
    //arrange
    const point = {
      id: 2,
      orgId: 1,
      userId: 1,
      amount: 300,
      created: '2024-08-25T09:06:58',
      updated: '2024-08-25T09:06:58',
    };
    const points = [point];

    jest.spyOn(mockPointRepository, 'find').mockReturnValue(points);

    //act
    const result = await service.getPoint();

    // assert
    expect(result).toEqual(points);
    expect(mockPointRepository.find).toBeCalled();
  });
});
